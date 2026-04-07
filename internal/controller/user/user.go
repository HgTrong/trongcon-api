package user

import (
	"errors"
	"net/http"
	"strconv"

	"trongcon-api/api/swagger"
	v1 "trongcon-api/api/user/v1"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.UserService
}

func NewController(svc service.UserService) *Controller {
	return &Controller{svc: svc}
}

// Create godoc
// @Summary Tạo user
// @Description Tạo user mới (mật khẩu được hash bcrypt).
// @Tags admin-users
// @Accept json
// @Produce json
// @Param body body v1.CreateReq true "Payload"
// @Success 201 {object} v1.CreateRes
// @Failure 400 {object} swagger.ErrBody
// @Failure 409 {object} swagger.ErrBody
// @Failure 500 {object} swagger.ErrBody
// @Router /admin/users [post]
func (c *Controller) Create(ctx *gin.Context) {
	var req v1.CreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Create(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

// List godoc
// @Summary Danh sách user (phân trang)
// @Tags admin-users
// @Produce json
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số bản ghi" default(10)
// @Success 200 {object} v1.ListUsersRes
// @Failure 400 {object} swagger.ErrBody
// @Failure 500 {object} swagger.ErrBody
// @Router /admin/users [get]
func (c *Controller) List(ctx *gin.Context) {
	var req v1.ListUsersReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.List(ctx.Request.Context(), &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// GetByID godoc
// @Summary Lấy user theo ID
// @Tags admin-users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} v1.GetRes
// @Failure 400 {object} swagger.ErrBody
// @Failure 404 {object} swagger.ErrBody
// @Failure 500 {object} swagger.ErrBody
// @Router /admin/users/{id} [get]
func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// Update godoc
// @Summary Cập nhật user
// @Tags admin-users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body v1.UpdateReq true "Payload"
// @Success 200 {object} v1.UpdateRes
// @Failure 400 {object} swagger.ErrBody
// @Failure 404 {object} swagger.ErrBody
// @Failure 409 {object} swagger.ErrBody
// @Failure 500 {object} swagger.ErrBody
// @Router /admin/users/{id} [put]
func (c *Controller) Update(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	var req v1.UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.Update(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// Delete godoc
// @Summary Xóa user (soft delete)
// @Tags admin-users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} v1.DeleteRes
// @Failure 400 {object} swagger.ErrBody
// @Failure 404 {object} swagger.ErrBody
// @Failure 500 {object} swagger.ErrBody
// @Router /admin/users/{id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, v1.DeleteRes{Status: "ok"})
}

func parseUintParam(ctx *gin.Context, name string) (uint, error) {
	s := ctx.Param(name)
	if s == "" {
		return 0, strconv.ErrRange
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailExists):
		ctx.JSON(http.StatusConflict, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrInvalidPayload):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
	}
}
