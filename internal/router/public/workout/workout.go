package workout

import (
	workoutctl "trongcon-api/internal/controller/workout"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *workoutctl.Controller, detailMW ...gin.HandlerFunc) {
	wg := g.Group("/workouts")
	{
		wg.GET("", c.List)
		detail := wg.Group("")
		detail.Use(detailMW...)
		detail.GET("/:id", c.GetByID)
	}
}
