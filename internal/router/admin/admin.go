package admin

import (
	articlectl "trongcon-api/internal/controller/article"
	categoryctl "trongcon-api/internal/controller/category"
	equipmentctl "trongcon-api/internal/controller/equipment"
	exercisectl "trongcon-api/internal/controller/exercise"
	foodctl "trongcon-api/internal/controller/food"
	gymctl "trongcon-api/internal/controller/gym"
	mealplanctl "trongcon-api/internal/controller/meal_plan"
	routinectl "trongcon-api/internal/controller/routine"
	workoutctl "trongcon-api/internal/controller/workout"
	musclectl "trongcon-api/internal/controller/muscle"
	toolsctl "trongcon-api/internal/controller/tools"
	uploadctl "trongcon-api/internal/controller/upload"
	userctl "trongcon-api/internal/controller/user"
	statsctl "trongcon-api/internal/controller/stats"
	planctl "trongcon-api/internal/controller/subscription_plan"
	subctl "trongcon-api/internal/controller/user_subscription"
	adminarticle "trongcon-api/internal/router/admin/article"
	adminbranch "trongcon-api/internal/router/admin/branch"
	admincategory "trongcon-api/internal/router/admin/category"
	adminequipment "trongcon-api/internal/router/admin/equipment"
	adminexercise "trongcon-api/internal/router/admin/exercise"
	adminfood "trongcon-api/internal/router/admin/food"
	adminmealplan "trongcon-api/internal/router/admin/meal_plan"
	adminroutine "trongcon-api/internal/router/admin/routine"
	admintrainer "trongcon-api/internal/router/admin/trainer"
	adminworkout "trongcon-api/internal/router/admin/workout"
	adminmuscle "trongcon-api/internal/router/admin/muscle"
	admintools "trongcon-api/internal/router/admin/tools"
	adminupload "trongcon-api/internal/router/admin/upload"
	adminuser "trongcon-api/internal/router/admin/user"
	adminstats "trongcon-api/internal/router/admin/stats"
	adminsubplan "trongcon-api/internal/router/admin/subscription_plan"
	adminusersub "trongcon-api/internal/router/admin/user_subscription"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	User             *userctl.Controller
	Category         *categoryctl.Controller
	Article          *articlectl.Controller
	Equipment        *equipmentctl.Controller
	Exercise         *exercisectl.Controller
	Food             *foodctl.Controller
	MealPlan         *mealplanctl.Controller
	Routine          *routinectl.Controller
	Workout          *workoutctl.Controller
	Muscle           *musclectl.Controller
	Tools            *toolsctl.Controller
	Upload           *uploadctl.Controller
	Stats            *statsctl.Controller
	Gym              *gymctl.Controller
	SubscriptionPlan *planctl.Controller
	UserSubscription *subctl.Controller
}

func Register(r *gin.RouterGroup, c Controllers) {
	adminuser.Register(r, c.User)
	admincategory.Register(r, c.Category)
	adminarticle.Register(r, c.Article)
	adminequipment.Register(r, c.Equipment)
	adminexercise.Register(r, c.Exercise)
	adminfood.Register(r, c.Food)
	adminmealplan.Register(r, c.MealPlan)
	adminroutine.Register(r, c.Routine)
	adminworkout.Register(r, c.Workout)
	adminmuscle.Register(r, c.Muscle)
	admintools.Register(r, c.Tools)
	adminupload.Register(r, c.Upload)
	adminstats.Register(r, c.Stats)
	if c.Gym != nil {
		adminbranch.Register(r, c.Gym)
		admintrainer.Register(r, c.Gym)
	}
	if c.SubscriptionPlan != nil {
		adminsubplan.Register(r, c.SubscriptionPlan)
	}
	if c.UserSubscription != nil {
		adminusersub.Register(r, c.UserSubscription)
	}
}
