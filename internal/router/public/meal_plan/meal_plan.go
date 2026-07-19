package meal_plan

import (
	mealplanctl "trongcon-api/internal/controller/meal_plan"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *mealplanctl.Controller, detailMW ...gin.HandlerFunc) {
	mpg := g.Group("/meal-plans")
	{
		mpg.GET("", c.ListPublic)
		detail := mpg.Group("")
		detail.Use(detailMW...)
		detail.GET("/:id", c.GetByIDPublic)
	}
}
