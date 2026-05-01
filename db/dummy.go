package db

import (
	"_discord_bot/models"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateLoginAttemptsDummyData(db *gorm.DB, month string) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	start, err := time.Parse("2006-01", month)
	if err != nil {
		return err
	}

	end := start.AddDate(0, 1, 0)

	rand.Seed(time.Now().UnixNano())

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {

		dailyCount := rand.Intn(49) + 2 // 2–50 per day

		for i := 0; i < dailyCount; i++ {
			t := randomTime(d, d.AddDate(0, 0, 1))

			attempt := models.LoginAttempt{
				UUID:      uuid.NewString(),
				UserUUID:  "",
				Success:   rand.Intn(100) < 70,
				IP:        randomIP(),
				UserAgent: randomUserAgent(),
				CreatedAt: t,
			}

			if err := db.Create(&attempt).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

/* ---------------- helpers ---------------- */

func randomTime(start, end time.Time) time.Time {
	delta := end.Sub(start)
	sec := rand.Int63n(int64(delta.Seconds()))
	return start.Add(time.Duration(sec) * time.Second)
}

func randomIP() string {
	return fmt.Sprintf(
		"192.168.%d.%d",
		rand.Intn(255),
		rand.Intn(255),
	)
}

func randomUserAgent() string {
	agents := []string{
		"Mozilla/5.0 Firefox",
		"Chrome/120",
		"Safari/17",
		"curl/8.0",
		"bot-scan",
	}
	return agents[rand.Intn(len(agents))]
}
