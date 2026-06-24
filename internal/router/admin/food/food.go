package food

import (
	foodctl "trongcon-api/internal/controller/food"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *foodctl.Controller) {
	fg := g.Group("/foods")
	{
		fg.POST("", c.Create)
		fg.GET("", c.List)
		fg.GET("/:id", c.GetByID)
		fg.PUT("/:id", c.Update)
		fg.DELETE("/:id", c.Delete)
	}
}
