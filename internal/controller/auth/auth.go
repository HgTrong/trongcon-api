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

func (c *Controller) ForgotPassword(ctx *gin.Context) {
	var req authv1.ForgotPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ForgotPassword(ctx.Request.Context(), &req)
	if err != nil {
		writeAuthErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) VerifyForgotOTP(ctx *gin.Context) {
	var req authv1.VerifyForgotOTPReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.VerifyForgotOTP(ctx.Request.Context(), &req)
	if err != nil {
		writeAuthErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ResetPassword(ctx *gin.Context) {
	var req authv1.ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.ResetPassword(ctx.Request.Context(), &req)
	if err != nil {
		writeAuthErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func writeAuthErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailExists):
		ctx.JSON(http.StatusConflict, swagger.ErrBody{Error: "Email đã được sử dụng"})
	case errors.Is(err, service.ErrInvalidCredentials):
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "Email hoặc mật khẩu không đúng"})
	case errors.Is(err, service.ErrNotSuper):
		ctx.JSON(http.StatusForbidden, swagger.ErrBody{Error: "Tài khoản không có quyền quản trị"})
	case errors.Is(err, service.ErrInvalidOTP):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "Mã OTP không hợp lệ hoặc đã hết hạn"})
	case errors.Is(err, service.ErrInvalidResetToken):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "Phiên đặt lại mật khẩu không hợp lệ hoặc đã hết hạn"})
	case errors.Is(err, service.ErrSMTPDisabled), errors.Is(err, service.ErrEmailTemplateNotFound), errors.Is(err, service.ErrEmailTemplateInactive):
		ctx.JSON(http.StatusServiceUnavailable, swagger.ErrBody{Error: "Hệ thống gửi email tạm thời không khả dụng"})
	default:
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "Yêu cầu không hợp lệ"})
	}
}
