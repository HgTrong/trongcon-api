package public

import (
	articlectl "trongcon-api/internal/controller/article"
	categoryctl "trongcon-api/internal/controller/category"
	equipmentctl "trongcon-api/internal/controller/equipment"
	exercisectl "trongcon-api/internal/controller/exercise"
	foodctl "trongcon-api/internal/controller/food"
	gymctl "trongcon-api/internal/controller/gym"
	mealplanctl "trongcon-api/internal/controller/meal_plan"
	musclectl "trongcon-api/internal/controller/muscle"
	routinectl "trongcon-api/internal/controller/routine"
	planctl "trongcon-api/internal/controller/subscription_plan"
	gymcommercectl "trongcon-api/internal/controller/gym_commerce"
	faqctl "trongcon-api/internal/controller/faq"
	toolsctl "trongcon-api/internal/controller/tools"
	workoutctl "trongcon-api/internal/controller/workout"
	"trongcon-api/internal/service"
	publicarticle "trongcon-api/internal/router/public/article"
	publicbranch "trongcon-api/internal/router/public/branch"
	publiccategory "trongcon-api/internal/router/public/category"
	publicequipment "trongcon-api/internal/router/public/equipment"
	publicexercise "trongcon-api/internal/router/public/exercise"
	publicfaq "trongcon-api/internal/router/public/faq"
	publicfood "trongcon-api/internal/router/public/food"
	publicgymcommerce "trongcon-api/internal/router/public/gym_commerce"
	publicmealplan "trongcon-api/internal/router/public/meal_plan"
	publicmuscle "trongcon-api/internal/router/public/muscle"
	publicroutine "trongcon-api/internal/router/public/routine"
	publicplan "trongcon-api/internal/router/public/subscription_plan"
	publictools "trongcon-api/internal/router/public/tools"
	publictrainer "trongcon-api/internal/router/public/trainer"
	publicworkout "trongcon-api/internal/router/public/workout"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	Exercise         *exercisectl.Controller
	Muscle           *musclectl.Controller
	Equipment        *equipmentctl.Controller
	Workout          *workoutctl.Controller
	Routine          *routinectl.Controller
	Food             *foodctl.Controller
	MealPlan         *mealplanctl.Controller
	Article          *articlectl.Controller
	Category         *categoryctl.Controller
	Tools            *toolsctl.Controller
	Gym              *gymctl.Controller
	SubscriptionPlan *planctl.Controller
	GymCommerce      *gymcommercectl.Controller
	FAQ              *faqctl.Controller
}

func Register(r *gin.RouterGroup, c Controllers, jwtSecret string, premium service.UserSubscriptionService) {
	// Catalog browse/detail is free (discovery for PTs + gym funnel).
	// Premium gates stay on member tools: sessions, food log, AI, clone/enroll, etc.
	_ = jwtSecret
	_ = premium

	publicexercise.Register(r, c.Exercise)
	publicmuscle.Register(r, c.Muscle)
	publicequipment.Register(r, c.Equipment)
	publicworkout.Register(r, c.Workout)
	publicroutine.Register(r, c.Routine)
	publicfood.Register(r, c.Food)
	publicmealplan.Register(r, c.MealPlan)
	publicarticle.Register(r, c.Article)
	publiccategory.Register(r, c.Category)
	publictools.Register(r, c.Tools)
	if c.SubscriptionPlan != nil {
		publicplan.Register(r, c.SubscriptionPlan)
	}
	if c.Gym != nil {
		publicbranch.Register(r, c.Gym)
		publictrainer.Register(r, c.Gym)
	}
	if c.GymCommerce != nil {
		publicgymcommerce.Register(r, c.GymCommerce)
	}
	if c.FAQ != nil {
		publicfaq.Register(r, c.FAQ)
	}
}
