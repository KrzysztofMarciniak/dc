package templates

import (
	"text/template"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	funcMap := template.FuncMap{
		"inc": func(i int) int { return i + 1 },
		"dec": func(i int) int { return i - 1 },
	}
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("templates/*.tmpl")
}
