package article

import (
	articlectl "trongcon-api/internal/controller/article"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *articlectl.Controller) {
	ag := g.Group("/articles")
	{
		ag.POST("", c.Create)
		ag.GET("", c.List)
		ag.GET("/:id", c.GetByID)
		ag.PUT("/:id", c.Update)
		ag.DELETE("/:id", c.Delete)
	}
}
