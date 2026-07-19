package http

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"trongcon-api/internal/config"
	authctl "trongcon-api/internal/controller/auth"
	aictl "trongcon-api/internal/controller/ai"
	savedctl "trongcon-api/internal/controller/saved_workout"
	foodlogctl "trongcon-api/internal/controller/food_log"
	mytrainctl "trongcon-api/internal/controller/my_train"
	sessionctl "trongcon-api/internal/controller/workout_session"
	enrollctl "trongcon-api/internal/controller/training_enrollment"
	subctl "trongcon-api/internal/controller/user_subscription"
	"trongcon-api/internal/http/handlers"
	"trongcon-api/internal/http/middleware"
	adminrouter "trongcon-api/internal/router/admin"
	publicrouter "trongcon-api/internal/router/public"
	userrouter "trongcon-api/internal/router/user"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, deps Deps) *gin.Engine {
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
			"docs":    "/swagger",
		})
	})

	registerSwaggerRoutes(router, cfg)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handlers.HealthCheck)

		userrouter.Register(v1.Group("/user"), userrouter.Controllers{
			Auth:         deps.Auth,
			User:         deps.Admin.User,
			Saved:        deps.Saved,
			FoodLog:      deps.FoodLog,
			MyTrain:      deps.MyTrain,
			Sessions:     deps.Sessions,
			Enrollment:   deps.Enrollment,
			AI:           deps.AI,
			Subscription: deps.Subscription,
		}, cfg.JWTSecret, deps.Premium)

		publicrouter.Register(v1, deps.Public, cfg.JWTSecret, deps.Premium)

		if deps.Subscription != nil {
			v1.POST("/webhooks/stripe", deps.Subscription.StripeWebhook)
		}

		admin := v1.Group("/admin")
		admin.POST("/login", deps.Auth.AdminLogin)
		admin.Use(middleware.RequireSuper(cfg.JWTSecret))
		adminrouter.Register(admin, deps.Admin)
	}

	return router
}

type Deps struct {
	Auth         *authctl.Controller
	Saved        *savedctl.Controller
	FoodLog      *foodlogctl.Controller
	MyTrain      *mytrainctl.Controller
	Sessions     *sessionctl.Controller
	Enrollment   *enrollctl.Controller
	AI           *aictl.Controller
	Subscription *subctl.Controller
	Premium      service.UserSubscriptionService
	Admin        adminrouter.Controllers
	Public       publicrouter.Controllers
}

func registerSwaggerRoutes(router *gin.Engine, cfg config.Config) {
	swaggerUI := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>SwaggerUI</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.10.5/swagger-ui.min.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.10.5/swagger-ui-bundle.js" crossorigin></script>
<script>
window.onload = () => {
  SwaggerUIBundle({
    url: '/api.json',
    dom_id: '#swagger-ui',
  });
};
</script>
</body>
</html>`

	router.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUI))
	})
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger")
	})
	router.GET("/api.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, buildOpenAPIDoc(router, cfg))
	})
}

func buildOpenAPIDoc(router *gin.Engine, cfg config.Config) gin.H {
	paths := map[string]gin.H{}
	tagSet := map[string]struct{}{}
	routes := router.Routes()
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	for _, r := range routes {
		if strings.HasPrefix(r.Path, "/swagger") || r.Path == "/api.json" || r.Path == "/" {
			continue
		}
		p := strings.ReplaceAll(r.Path, ":", "{")
		if strings.Contains(p, "{") {
			p = closePathParams(p)
		}
		if _, ok := paths[p]; !ok {
			paths[p] = gin.H{}
		}
		method := strings.ToLower(r.Method)
		tag := tagForPath(r.Path)
		tagSet[tag] = struct{}{}
		paths[p][method] = gin.H{
			"summary":     prettySummary(r.Handler),
			"operationId": handlerOperationID(r.Handler, r.Method, p),
			"tags":        []string{tag},
			"responses": gin.H{
				"default": gin.H{"description": "Default response"},
			},
		}
	}
	tags := make([]gin.H, 0, len(tagSet))
	tagNames := make([]string, 0, len(tagSet))
	for name := range tagSet {
		tagNames = append(tagNames, name)
	}
	sort.Strings(tagNames)
	for _, name := range tagNames {
		tags = append(tags, gin.H{"name": name})
	}

	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "TrongCon API",
			"version":     "1.0.0",
			"description": "Runtime OpenAPI generated from registered routes",
		},
		"servers": []gin.H{
			{"url": fmt.Sprintf("http://localhost:%s", cfg.Port)},
		},
		"tags":  tags,
		"paths": paths,
	}
}

func closePathParams(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, "{") && !strings.HasSuffix(part, "}") {
			parts[i] = part + "}"
		}
	}
	return strings.Join(parts, "/")
}

func handlerShortName(h string) string {
	parts := strings.Split(h, ".")
	if len(parts) == 0 {
		return h
	}
	return parts[len(parts)-1]
}

func prettySummary(handler string) string {
	name := handlerShortName(handler)
	name = strings.TrimSuffix(name, "-fm")
	if name == "" {
		return "Handler"
	}
	return name
}

func tagForPath(path string) string {
	if path == "/api/v1/health" {
		return "health"
	}
	if strings.HasPrefix(path, "/api/v1/user/") {
		return "user-auth"
	}
	if strings.HasPrefix(path, "/api/v1/admin/login") {
		return "admin-auth"
	}
	if strings.HasPrefix(path, "/api/v1/admin/") {
		rest := strings.TrimPrefix(path, "/api/v1/admin/")
		parts := strings.Split(rest, "/")
		if len(parts) > 0 && parts[0] != "" {
			return "admin-" + parts[0]
		}
		return "admin"
	}
	if strings.HasPrefix(path, "/api/v1/") {
		rest := strings.TrimPrefix(path, "/api/v1/")
		parts := strings.Split(rest, "/")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}
	return "default"
}

func handlerOperationID(handler, method, path string) string {
	clean := strings.ReplaceAll(path, "/", "_")
	clean = strings.ReplaceAll(clean, "{", "")
	clean = strings.ReplaceAll(clean, "}", "")
	clean = strings.Trim(clean, "_")
	return strings.ToLower(method) + "_" + clean + "_" + handlerShortName(handler)
}
