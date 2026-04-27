package templates

import (
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	r.LoadHTMLGlob("templates/*.tmpl")
}
