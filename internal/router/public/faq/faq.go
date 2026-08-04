package faq

import (
	faqctl "trongcon-api/internal/controller/faq"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *faqctl.Controller) {
	fg := g.Group("/faqs")
	{
		fg.GET("", c.ListPublic)
		fg.GET("/:id", c.GetByIDPublic)
	}
}
