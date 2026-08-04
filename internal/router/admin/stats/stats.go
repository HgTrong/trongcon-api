package stats

import (
	statsctl "trongcon-api/internal/controller/stats"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *statsctl.Controller) {
	g.GET("/stats/overview", c.Overview)
	g.GET("/stats/top-trainers", c.TopTrainers)
}
