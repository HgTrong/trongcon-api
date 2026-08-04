package emailtemplate

import (
	etctl "trongcon-api/internal/controller/email_template"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *etctl.Controller) {
	eg := g.Group("/email-templates")
	{
		eg.GET("", c.List)
		eg.POST("", c.Create)
		eg.POST("/preview", c.Preview)
		eg.GET("/:id", c.GetByID)
		eg.PUT("/:id", c.Update)
		eg.DELETE("/:id", c.Delete)
		eg.POST("/:id/test-send", c.TestSend)
	}
}
