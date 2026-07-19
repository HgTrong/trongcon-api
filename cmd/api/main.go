// Package main khởi chạy API TrongCon.
// @title TrongCon API
// @version 1.0
// @description REST API: auth, admin CRUD, user signup/login. Swagger: /swagger/index.html
// @termsOfService http://swagger.io/terms/

// @contact.name TrongCon
// @contact.url https://github.com
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"trongcon-api/internal/config"
	articlectl "trongcon-api/internal/controller/article"
	authctl "trongcon-api/internal/controller/auth"
	categoryctl "trongcon-api/internal/controller/category"
	equipmentctl "trongcon-api/internal/controller/equipment"
	exercisectl "trongcon-api/internal/controller/exercise"
	foodctl "trongcon-api/internal/controller/food"
	mealplanctl "trongcon-api/internal/controller/meal_plan"
	routinectl "trongcon-api/internal/controller/routine"
	workoutctl "trongcon-api/internal/controller/workout"
	musclectl "trongcon-api/internal/controller/muscle"
	toolsctl "trongcon-api/internal/controller/tools"
	uploadctl "trongcon-api/internal/controller/upload"
	savedworkoutctl "trongcon-api/internal/controller/saved_workout"
	foodlogctl "trongcon-api/internal/controller/food_log"
	mytrainctl "trongcon-api/internal/controller/my_train"
	sessionctl "trongcon-api/internal/controller/workout_session"
	enrollctl "trongcon-api/internal/controller/training_enrollment"
	aictl "trongcon-api/internal/controller/ai"
	gymctl "trongcon-api/internal/controller/gym"
	planctl "trongcon-api/internal/controller/subscription_plan"
	subctl "trongcon-api/internal/controller/user_subscription"
	userctl "trongcon-api/internal/controller/user"
	statsctl "trongcon-api/internal/controller/stats"
	httpserver "trongcon-api/internal/http"
	oaiclient "trongcon-api/internal/openai"
	"trongcon-api/internal/repository"
	adminrouter "trongcon-api/internal/router/admin"
	publicrouter "trongcon-api/internal/router/public"
	"trongcon-api/internal/service"
	"trongcon-api/internal/storage/postgres"
)

func main() {
	loadDotEnv()

	cfg := config.Load()
	db := postgres.GetDatabaseConnection()
	userRepo := repository.NewUserRepository(db.Connection)
	roleRepo := repository.NewRoleRepository(db.Connection)
	categoryRepo := repository.NewCategoryRepository(db.Connection)
	articleRepo := repository.NewArticleRepository(db.Connection)
	equipmentRepo := repository.NewEquipmentRepository(db.Connection)
	exerciseRepo := repository.NewExerciseRepository(db.Connection)
	foodRepo := repository.NewFoodRepository(db.Connection)
	mealPlanRepo := repository.NewMealPlanRepository(db.Connection)
	routineRepo := repository.NewRoutineRepository(db.Connection)
	workoutRepo := repository.NewWorkoutRepository(db.Connection)
	savedWorkoutRepo := repository.NewSavedWorkoutRepository(db.Connection)
	foodLogRepo := repository.NewFoodLogRepository(db.Connection)
	sessionRepo := repository.NewWorkoutSessionRepository(db.Connection)
	enrollmentRepo := repository.NewTrainingEnrollmentRepository(db.Connection)
	muscleRepo := repository.NewMuscleRepository(db.Connection)
	aiChatRepo := repository.NewAiChatRepository(db.Connection)
	branchRepo := repository.NewGymBranchRepository(db.Connection)
	trainerRepo := repository.NewTrainerProfileRepository(db.Connection)

	uploadSvc := service.NewUploadService(cfg.S3)
	userSvc := service.NewUserService(userRepo, roleRepo, uploadSvc)
	categorySvc := service.NewCategoryService(categoryRepo)
	articleSvc := service.NewArticleService(articleRepo, categoryRepo, userRepo)
	equipmentSvc := service.NewEquipmentService(equipmentRepo)
	exerciseSvc := service.NewExerciseService(exerciseRepo, equipmentRepo, muscleRepo)
	foodSvc := service.NewFoodService(foodRepo)
	mealPlanSvc := service.NewMealPlanService(mealPlanRepo, foodRepo, userRepo, trainerRepo)
	routineSvc := service.NewRoutineService(routineRepo, workoutRepo, userRepo, trainerRepo)
	workoutSvc := service.NewWorkoutService(workoutRepo, exerciseRepo, userRepo, trainerRepo)
	savedWorkoutSvc := service.NewSavedWorkoutService(savedWorkoutRepo, workoutRepo)
	macroSvc := service.NewMacroService()
	foodLogSvc := service.NewFoodLogService(foodLogRepo, foodRepo, macroSvc)
	myTrainSvc := service.NewMyTrainService(workoutRepo, exerciseRepo, routineRepo)
	sessionSvc := service.NewWorkoutSessionService(sessionRepo, workoutRepo, exerciseRepo)
	enrollmentSvc := service.NewTrainingEnrollmentService(enrollmentRepo, routineRepo, workoutRepo, sessionRepo, sessionSvc)
	muscleSvc := service.NewMuscleService(muscleRepo)
	tdeeSvc := service.NewTDEEService()
	oneRepMaxSvc := service.NewOneRepMaxService()
	statsSvc := service.NewStatsService(db.Connection)
	openaiClient := oaiclient.NewClient(cfg.OpenAI)
	aiCoachSvc := service.NewAICoachService(openaiClient, foodRepo, exerciseRepo, muscleRepo, mealPlanSvc, myTrainSvc, aiChatRepo)
	gymSvc := service.NewGymService(branchRepo, trainerRepo, userRepo, roleRepo, workoutRepo, routineRepo, mealPlanRepo)

	planRepo := repository.NewSubscriptionPlanRepository(db.Connection)
	userSubRepo := repository.NewUserSubscriptionRepository(db.Connection)
	paymentHistoryRepo := repository.NewPaymentHistoryRepository(db.Connection)
	paypalSvc, err := service.NewPayPalService(cfg.PayPal)
	if err != nil {
		log.Printf("PayPal init warning: %v — using mock", err)
		cfg.PayPal.TestMode = "mock"
		paypalSvc, _ = service.NewPayPalService(cfg.PayPal)
	}
	planSvc := service.NewSubscriptionPlanService(planRepo)
	stripeSvc := service.NewStripeService(cfg.Stripe)
	userSubSvc := service.NewUserSubscriptionService(userSubRepo, planRepo, paymentHistoryRepo, userRepo, paypalSvc, stripeSvc)

	authSvc := service.NewAuthService(userRepo, roleRepo, cfg.JWTSecret, cfg.JWTExpiration)

	userController := userctl.NewController(userSvc)
	authController := authctl.NewController(authSvc)
	categoryController := categoryctl.NewController(categorySvc)
	articleController := articlectl.NewController(articleSvc)
	equipmentController := equipmentctl.NewController(equipmentSvc)
	exerciseController := exercisectl.NewController(exerciseSvc)
	foodController := foodctl.NewController(foodSvc)
	mealPlanController := mealplanctl.NewController(mealPlanSvc)
	routineController := routinectl.NewController(routineSvc)
	workoutController := workoutctl.NewController(workoutSvc)
	savedWorkoutController := savedworkoutctl.NewController(savedWorkoutSvc)
	foodLogController := foodlogctl.NewController(foodLogSvc)
	myTrainController := mytrainctl.NewController(myTrainSvc)
	sessionController := sessionctl.NewController(sessionSvc)
	enrollmentController := enrollctl.NewController(enrollmentSvc)
	aiController := aictl.NewController(aiCoachSvc, mealPlanSvc)
	gymController := gymctl.NewController(gymSvc)
	muscleController := musclectl.NewController(muscleSvc)
	toolsController := toolsctl.NewController(tdeeSvc, macroSvc, oneRepMaxSvc)
	uploadController := uploadctl.NewController(uploadSvc)
	statsController := statsctl.NewController(statsSvc)
	planController := planctl.NewController(planSvc)
	userSubController := subctl.NewController(userSubSvc, cfg.Stripe)

	router := httpserver.NewRouter(cfg, httpserver.Deps{
		Auth:         authController,
		Saved:        savedWorkoutController,
		FoodLog:      foodLogController,
		MyTrain:      myTrainController,
		Sessions:     sessionController,
		Enrollment:   enrollmentController,
		AI:           aiController,
		Subscription: userSubController,
		Premium:      userSubSvc,
		Admin: adminrouter.Controllers{
			User:             userController,
			Category:         categoryController,
			Article:          articleController,
			Equipment:        equipmentController,
			Exercise:         exerciseController,
			Food:             foodController,
			MealPlan:         mealPlanController,
			Routine:          routineController,
			Workout:          workoutController,
			Muscle:           muscleController,
			Tools:            toolsController,
			Upload:           uploadController,
			Stats:            statsController,
			Gym:              gymController,
			SubscriptionPlan: planController,
			UserSubscription: userSubController,
		},
		Public: publicrouter.Controllers{
			Exercise:         exerciseController,
			Muscle:           muscleController,
			Equipment:        equipmentController,
			Workout:          workoutController,
			Routine:          routineController,
			Food:             foodController,
			MealPlan:         mealPlanController,
			Article:          articleController,
			Category:         categoryController,
			Tools:            toolsController,
			Gym:              gymController,
			SubscriptionPlan: planController,
		},
	})

	log.Printf("API listening on :%s — Swagger: http://localhost:%s/swagger/index.html", cfg.Port, cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

// loadDotEnv loads .env from cwd or parent dirs (so `go run` from cmd/api still works).
func loadDotEnv() {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 4; i++ {
			candidates = append(candidates, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	for _, p := range candidates {
		if err := godotenv.Load(p); err == nil {
			return
		}
	}
}
