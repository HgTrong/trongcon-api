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
	revenuectl "trongcon-api/internal/controller/revenue"
	planctl "trongcon-api/internal/controller/subscription_plan"
	subctl "trongcon-api/internal/controller/user_subscription"
	gymcommercectl "trongcon-api/internal/controller/gym_commerce"
	emailtemplatectl "trongcon-api/internal/controller/email_template"
	faqctl "trongcon-api/internal/controller/faq"
	adminarticle "trongcon-api/internal/router/admin/article"
	adminbranch "trongcon-api/internal/router/admin/branch"
	admincategory "trongcon-api/internal/router/admin/category"
	adminemailtemplate "trongcon-api/internal/router/admin/email_template"
	adminequipment "trongcon-api/internal/router/admin/equipment"
	adminexercise "trongcon-api/internal/router/admin/exercise"
	adminfaq "trongcon-api/internal/router/admin/faq"
	adminfood "trongcon-api/internal/router/admin/food"
	admingymcommerce "trongcon-api/internal/router/admin/gym_commerce"
	adminmealplan "trongcon-api/internal/router/admin/meal_plan"
	adminroutine "trongcon-api/internal/router/admin/routine"
	admintrainer "trongcon-api/internal/router/admin/trainer"
	adminworkout "trongcon-api/internal/router/admin/workout"
	adminmuscle "trongcon-api/internal/router/admin/muscle"
	admintools "trongcon-api/internal/router/admin/tools"
	adminupload "trongcon-api/internal/router/admin/upload"
	adminuser "trongcon-api/internal/router/admin/user"
	adminstats "trongcon-api/internal/router/admin/stats"
	adminrevenue "trongcon-api/internal/router/admin/revenue"
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
	Revenue          *revenuectl.Controller
	Gym              *gymctl.Controller
	SubscriptionPlan *planctl.Controller
	UserSubscription *subctl.Controller
	GymCommerce      *gymcommercectl.Controller
	EmailTemplate    *emailtemplatectl.Controller
	FAQ              *faqctl.Controller
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
	if c.Revenue != nil {
		adminrevenue.Register(r, c.Revenue)
	}
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
	if c.GymCommerce != nil {
		admingymcommerce.Register(r, c.GymCommerce)
	}
	if c.EmailTemplate != nil {
		adminemailtemplate.Register(r, c.EmailTemplate)
	}
	if c.FAQ != nil {
		adminfaq.Register(r, c.FAQ)
	}
}
