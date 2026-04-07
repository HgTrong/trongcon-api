package auth

import (
	"errors"
	"net/http"

	authv1 "trongcon-api/api/auth/v1"
	"trongcon-api/api/swagger"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.AuthService
}

func NewController(svc service.AuthService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) AdminLogin(ctx *gin.Context) {
	var req authv1.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.AdminLogin(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UserLogin(ctx *gin.Context) {
	var req authv1.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UserLogin(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeAuthErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) Signup(ctx *gin.Context) {
	var req authv1.SignupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Signup(ctx.Request.Context(), &req)
	if err != nil {
		writeAuthErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func writeAuthErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailExists):
		ctx.JSON(http.StatusConflict, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrNotSuper):
		ctx.JSON(http.StatusForbidden, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
	}
}
