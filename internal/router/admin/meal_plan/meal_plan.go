package meal_plan

import (
	mealplanctl "trongcon-api/internal/controller/meal_plan"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, c *mealplanctl.Controller) {
	mpg := g.Group("/meal-plans")
	{
		mpg.POST("", c.Create)
		mpg.GET("", c.List)
		mpg.GET("/:id", c.GetByID)
		mpg.PUT("/:id", c.Update)
		mpg.DELETE("/:id", c.Delete)
	}
}
