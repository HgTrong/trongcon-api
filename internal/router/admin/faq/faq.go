package faq

import (
	faqctl "trongcon-api/internal/controller/faq"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *faqctl.Controller) {
	fg := g.Group("/faqs")
	{
		fg.POST("", c.Create)
		fg.GET("", c.List)
		fg.GET("/:id", c.GetByID)
		fg.PUT("/:id", c.Update)
		fg.DELETE("/:id", c.Delete)
	}
}
