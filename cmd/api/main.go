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

	"github.com/joho/godotenv"
	_ "trongcon-api/docs"
	"trongcon-api/internal/config"
	articlectl "trongcon-api/internal/controller/article"
	authctl "trongcon-api/internal/controller/auth"
	categoryctl "trongcon-api/internal/controller/category"
	equipmentctl "trongcon-api/internal/controller/equipment"
	uploadctl "trongcon-api/internal/controller/upload"
	userctl "trongcon-api/internal/controller/user"
	httpserver "trongcon-api/internal/http"
	"trongcon-api/internal/repository"
	adminrouter "trongcon-api/internal/router/admin"
	"trongcon-api/internal/service"
	"trongcon-api/internal/storage/postgres"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	db := postgres.GetDatabaseConnection()
	userRepo := repository.NewUserRepository(db.Connection)
	roleRepo := repository.NewRoleRepository(db.Connection)
	categoryRepo := repository.NewCategoryRepository(db.Connection)
	articleRepo := repository.NewArticleRepository(db.Connection)
	equipmentRepo := repository.NewEquipmentRepository(db.Connection)

	userSvc := service.NewUserService(userRepo, roleRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	articleSvc := service.NewArticleService(articleRepo, categoryRepo, userRepo)
	equipmentSvc := service.NewEquipmentService(equipmentRepo)
	uploadSvc := service.NewUploadService(cfg.S3)

	authSvc := service.NewAuthService(userRepo, roleRepo, cfg.JWTSecret, cfg.JWTExpiration)

	userController := userctl.NewController(userSvc)
	authController := authctl.NewController(authSvc)
	categoryController := categoryctl.NewController(categorySvc)
	articleController := articlectl.NewController(articleSvc)
	equipmentController := equipmentctl.NewController(equipmentSvc)
	uploadController := uploadctl.NewController(uploadSvc)

	router := httpserver.NewRouter(cfg, authController, adminrouter.Controllers{
		User:     userController,
		Category: categoryController,
		Article:  articleController,
		Equipment: equipmentController,
		Upload:   uploadController,
	})

	log.Printf("API listening on :%s — Swagger: http://localhost:%s/swagger/index.html", cfg.Port, cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
