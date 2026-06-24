package tools

import (
	toolsctl "trongcon-api/internal/controller/tools"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *toolsctl.Controller) {
	tg := g.Group("/tools")
	{
		tg.POST("/tdee", c.CalculateTDEE)
		tg.POST("/macros", c.CalculateMacro)
		tg.POST("/one-rep-max", c.CalculateOneRepMax)
	}
}
