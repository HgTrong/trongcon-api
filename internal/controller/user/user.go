package user

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"trongcon-api/api/swagger"
	v1 "trongcon-api/api/user/v1"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.UserService
}

func NewController(svc service.UserService) *Controller {
	return &Controller{svc: svc}
}

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

func (c *Controller) GetMe(ctx *gin.Context) {
	id, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateMe(ctx *gin.Context) {
	id, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req v1.ProfileUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	res, err := c.svc.UpdateProfile(ctx.Request.Context(), id, &req)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ChangePassword(ctx *gin.Context) {
	id, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req v1.ChangePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	if err := c.svc.ChangePassword(ctx.Request.Context(), id, req.CurrentPassword, req.NewPassword); err != nil {
		writePasswordErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, v1.ChangePasswordRes{Status: "ok"})
}

const maxAvatarBytes = 5 << 20

var allowedAvatarExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
	".gif":  {},
}

func (c *Controller) UpdateAvatar(ctx *gin.Context) {
	id, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}

	res, err := c.uploadAvatarForUser(ctx, id)
	if err != nil {
		if status, msg, ok := avatarErrStatus(err); ok {
			ctx.JSON(status, swagger.ErrBody{Error: msg})
			return
		}
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) UpdateUserAvatar(ctx *gin.Context) {
	id, err := parseUintParam(ctx, "id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid id"})
		return
	}
	res, err := c.uploadAvatarForUser(ctx, id)
	if err != nil {
		if status, msg, ok := avatarErrStatus(err); ok {
			ctx.JSON(status, swagger.ErrBody{Error: msg})
			return
		}
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) uploadAvatarForUser(ctx *gin.Context, id uint) (*v1.UpdateAvatarRes, error) {
	file, err := ctx.FormFile("file")
	if err != nil {
		return nil, avatarInputErr{status: http.StatusBadRequest, msg: "missing multipart field file"}
	}
	if file.Size > maxAvatarBytes {
		return nil, avatarInputErr{status: http.StatusBadRequest, msg: "file too large (max 5MB)"}
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedAvatarExt[ext]; !ok {
		return nil, avatarInputErr{status: http.StatusBadRequest, msg: "invalid image type: use jpg, png, webp, or gif"}
	}

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	body, err := io.ReadAll(io.LimitReader(src, maxAvatarBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxAvatarBytes {
		return nil, avatarInputErr{status: http.StatusBadRequest, msg: "file too large (max 5MB)"}
	}

	ct := file.Header.Get("Content-Type")
	if ct == "" {
		switch ext {
		case ".png":
			ct = "image/png"
		case ".webp":
			ct = "image/webp"
		case ".gif":
			ct = "image/gif"
		default:
			ct = "image/jpeg"
		}
	}

	return c.svc.UpdateAvatar(ctx.Request.Context(), id, file.Filename, bytes.NewReader(body), ct)
}

type avatarInputErr struct {
	status int
	msg    string
}

func (e avatarInputErr) Error() string { return e.msg }

func avatarErrStatus(err error) (int, string, bool) {
	var input avatarInputErr
	if errors.As(err, &input) {
		return input.status, input.msg, true
	}
	return 0, "", false
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
		ctx.JSON(http.StatusConflict, swagger.ErrBody{Error: "Email đã được sử dụng"})
	case errors.Is(err, service.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: "Không tìm thấy người dùng"})
	case errors.Is(err, service.ErrInvalidPayload):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "Dữ liệu không hợp lệ"})
	case errors.Is(err, service.ErrS3NotConfigured):
		ctx.JSON(http.StatusServiceUnavailable, swagger.ErrBody{Error: "Dịch vụ tải ảnh chưa cấu hình"})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: "Có lỗi xảy ra. Vui lòng thử lại"})
	}
}

func writePasswordErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWrongPassword):
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "Mật khẩu hiện tại không đúng"})
	case errors.Is(err, service.ErrUserNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: "Không tìm thấy người dùng"})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: "Có lỗi xảy ra. Vui lòng thử lại"})
	}
}
