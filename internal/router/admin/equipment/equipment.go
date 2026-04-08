package equipment

import (
	equipmentctl "trongcon-api/internal/controller/equipment"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *equipmentctl.Controller) {
	eg := g.Group("/equipments")
	{
		eg.POST("", c.Create)
		eg.GET("", c.List)
		eg.GET("/:id", c.GetByID)
		eg.PUT("/:id", c.Update)
		eg.DELETE("/:id", c.Delete)
	}
}
