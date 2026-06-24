package exercise

import (
	exercisectl "trongcon-api/internal/controller/exercise"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *exercisectl.Controller) {
	eg := g.Group("/exercises")
	{
		eg.POST("", c.Create)
		eg.GET("", c.List)
		eg.GET("/:id", c.GetByID)
		eg.PUT("/:id", c.Update)
		eg.DELETE("/:id", c.Delete)
	}
}
