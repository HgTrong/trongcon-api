package routine

import (
	routinectl "trongcon-api/internal/controller/routine"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *routinectl.Controller) {
	rg := g.Group("/routines")
	{
		rg.POST("", c.Create)
		rg.GET("", c.List)
		rg.GET("/:id", c.GetByID)
		rg.PUT("/:id", c.Update)
		rg.DELETE("/:id", c.Delete)
	}
}
