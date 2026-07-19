package routine

import (
	routinectl "trongcon-api/internal/controller/routine"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *routinectl.Controller, detailMW ...gin.HandlerFunc) {
	rg := g.Group("/routines")
	{
		rg.GET("", c.ListPublic)
		detail := rg.Group("")
		detail.Use(detailMW...)
		detail.GET("/:id", c.GetByIDPublic)
	}
}
