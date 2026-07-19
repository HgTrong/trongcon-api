package category

import (
	categoryctl "trongcon-api/internal/controller/category"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *categoryctl.Controller) {
	cg := g.Group("/categories")
	{
		cg.GET("", c.ListPublic)
		cg.GET("/:id", c.GetByIDPublic)
	}
}
