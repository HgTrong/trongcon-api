package http

import (
	"net/http"

	"trongcon-api/internal/config"
	authctl "trongcon-api/internal/controller/auth"
	"trongcon-api/internal/http/handlers"
	"trongcon-api/internal/http/middleware"
	adminrouter "trongcon-api/internal/router/admin"
	userrouter "trongcon-api/internal/router/user"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(cfg config.Config, authCtrl *authctl.Controller, adminCtrls adminrouter.Controllers) *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "trongcon-api is running",
			"docs":    "/swagger/index.html",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handlers.HealthCheck)

		userrouter.Register(v1.Group("/user"), authCtrl)

		admin := v1.Group("/admin")
		admin.POST("/login", authCtrl.AdminLogin)
		admin.Use(middleware.RequireSuper(cfg.JWTSecret))
		adminrouter.Register(admin, adminCtrls)
	}

	return router
}
