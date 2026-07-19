package muscle

import (
	musclectl "trongcon-api/internal/controller/muscle"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *musclectl.Controller) {
	mg := g.Group("/muscles")
	{
		mg.GET("", c.List)
		mg.GET("/:id", c.GetByID)
	}
}
