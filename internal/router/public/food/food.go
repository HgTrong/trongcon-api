package food

import (
	foodctl "trongcon-api/internal/controller/food"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *foodctl.Controller) {
	fg := g.Group("/foods")
	{
		fg.GET("", c.List)
		fg.GET("/:id", c.GetByID)
	}
}
