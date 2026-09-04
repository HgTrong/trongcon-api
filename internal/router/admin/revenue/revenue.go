package revenue

import (
	revenuectl "trongcon-api/internal/controller/revenue"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *revenuectl.Controller) {
	g.GET("/revenue/overview", c.Overview)
	g.GET("/revenue/payments", c.Payments)
	g.GET("/revenue/pt-leaderboard", c.PTLeaderboard)
	g.GET("/dashboard/today-activity", c.TodayActivity)
}
