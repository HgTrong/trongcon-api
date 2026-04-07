package category

import (
	categoryctl "trongcon-api/internal/controller/category"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *categoryctl.Controller) {
	cg := g.Group("/categories")
	{
		cg.POST("", c.Create)
		cg.GET("", c.List)
		cg.GET("/:id", c.GetByID)
		cg.PUT("/:id", c.Update)
		cg.DELETE("/:id", c.Delete)
	}
}
