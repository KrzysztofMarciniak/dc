package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"_discord_bot/utils"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func computeDBHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "error"
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "error"
	}

	return hex.EncodeToString(h.Sum(nil))
}

func getDBSize(path string) string {
	fi, err := os.Stat(path)
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%d KB", fi.Size()/1024)
}

func ShowAdminPanel(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	page := 1
	pageSize := 5

	if p := c.Query("page"); p != "" {
		_, _ = fmt.Sscanf(p, "%d", &page)
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	attempts, total, err := models.GetLoginAttemptsWithUser(db.DB, pageSize, offset)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin-panel.tmpl", gin.H{
			"title": "Admin Panel",
			"error": "Failed to load login attempts",
		})
		return
	}

	for i := range attempts {
		attempts[i].UserAgentPretty = utils.ParseUserAgent(attempts[i].UserAgent)
		attempts[i].CreatedAtPretty = attempts[i].CreatedAt.Format("2006-01-02 15:04:05")
	}

	serverTime := time.Now().Format(time.RFC1123)
	uptime := time.Since(startTime).String()
	totalRequests := middleware.TotalRequests()

	dbStatus := "OK"
	if err := db.DB.Raw("SELECT 1").Scan(new(int)).Error; err != nil {
		dbStatus = "DOWN"
	}

	clientIP := c.ClientIP()
	userAgentRaw := c.Request.UserAgent()
	userAgent := utils.ParseUserAgent(userAgentRaw)

	dbSize := getDBSize(db.DBPath)
	dbHash := computeDBHash(db.DBPath)

	totalPages := 0
	if total > 0 {
		totalPages = (int(total) + pageSize - 1) / pageSize
	}

	csrf, err := models.NewCSRFToken(db.Secret)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin-panel.tmpl", gin.H{
			"title": "Admin Panel",
			"error": "Failed to generate CSRF token",
		})
		return
	}
	user, _ := models.GetUserByUUID(db.DB, auth.UUID)
	timeLeft := models.PasswordTimeLeft(user)

	var passwordExpiry string

	if timeLeft <= 0 {
		passwordExpiry = "expired"
	} else {
		totalMinutes := int(timeLeft.Minutes())

		days := totalMinutes / (60 * 24)
		remainingMinutes := totalMinutes % (60 * 24)

		hours := remainingMinutes / 60
		minutes := remainingMinutes % 60

		passwordExpiry = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}

	c.HTML(http.StatusOK, "admin-panel.tmpl", gin.H{
		"title":         "Admin Panel",
		"auth":          auth,
		"isLogged":      true,
		"loginAttempts": attempts,
		"timeLeft":      passwordExpiry,

		"page":       page,
		"totalPages": totalPages,
		"hasPrev":    page > 1,
		"hasNext":    page < totalPages,

		"serverTime":    serverTime,
		"uptime":        uptime,
		"totalRequests": totalRequests,
		"dbStatus":      dbStatus,

		"clientIP":  clientIP,
		"userAgent": userAgent,
		"dbSize":    dbSize,
		"dbHash":    dbHash,
		"csrf":      csrf,

		"success": c.Query("success"),
		"error":   c.Query("error"),
		"message": c.Query("message"),
	})
}
func DownloadDatabase(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	logEntry := models.InternalLog{
		ActorUUID: auth.UUID,
		ActorType: "user",
		Action:    "DOWNLOAD_DATABASE_BACKUP",
		TargetID:  "SYSTEM_DB",
		OldValue:  "",
		NewValue:  db.DBPath,
		IPAddress: c.ClientIP(),
	}

	if err := models.RecordInternalLog(db.DB, logEntry); err != nil {
		fmt.Printf("Failed to log DB download: %v\n", err)
	}

	filename := fmt.Sprintf("database-%s.db", time.Now().Format("2006-01-02"))
	c.FileAttachment(db.DBPath, filename)
}
func UploadDatabase(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	csrf := c.PostForm("csrf")
	valid, err := models.ValidateCSRFToken(csrf, db.Secret)
	if err != nil || !valid {
		c.Redirect(http.StatusFound, "/admin-panel?error=Invalid+csrf+token")
		return
	}

	file, err := c.FormFile("database")
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-panel?error=No+file+uploaded")
		return
	}

	if filepath.Ext(file.Filename) != ".db" {
		c.Redirect(http.StatusFound, "/admin-panel?error=Invalid+file+type")
		return
	}

	tmp := db.DBPath + ".tmp"

	if err := c.SaveUploadedFile(file, tmp); err != nil {
		c.Redirect(http.StatusFound, "/admin-panel?error=Upload+failed")
		return
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-panel?error=Read+failed")
		return
	}

	if err := os.WriteFile(db.DBPath, data, 0644); err != nil {
		c.Redirect(http.StatusFound, "/admin-panel?error=Write+failed")
		return
	}

	_ = os.Remove(tmp)

	logEntry := models.InternalLog{
		ActorUUID: auth.UUID,
		ActorType: "user",
		Action:    "UPLOAD_DATABASE_OVERWRITE",
		TargetID:  "SYSTEM_DB",
		OldValue:  "EXISTING_DATABASE_FILE",
		NewValue:  file.Filename,
		IPAddress: c.ClientIP(),
	}

	_ = models.RecordInternalLog(db.DB, logEntry)

	c.Redirect(http.StatusFound, "/admin-panel?success=Database+uploaded")
}
