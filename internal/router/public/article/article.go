package article

import (
	articlectl "trongcon-api/internal/controller/article"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *articlectl.Controller, detailMW ...gin.HandlerFunc) {
	ag := g.Group("/articles")
	{
		ag.GET("", c.ListPublic)
		detail := ag.Group("")
		detail.Use(detailMW...)
		detail.GET("/:slug", c.GetBySlug)
	}
}
