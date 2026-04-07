package user

import (
	authctl "trongcon-api/internal/controller/auth"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup, authCtrl *authctl.Controller) {
	g.POST("/signup", authCtrl.Signup)
	g.POST("/login", authCtrl.UserLogin)
}
