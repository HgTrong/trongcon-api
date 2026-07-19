package exercise

import (
	exercisectl "trongcon-api/internal/controller/exercise"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *exercisectl.Controller) {
	eg := g.Group("/exercises")
	{
		eg.GET("", c.ListPublic)
		eg.GET("/id/:id", c.GetByIDPublic)
		eg.POST("/:slug/view", c.IncrementViews)
		eg.GET("/:slug", c.GetBySlug)
	}
}
