package content_share

import (
	"errors"
	"net/http"
	"strconv"

	contentsharev1 "trongcon-api/api/content_share/v1"
	"trongcon-api/api/swagger"
	"trongcon-api/internal/http/middleware"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	svc service.ContentShareService
}

func NewController(svc service.ContentShareService) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Share(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	var req contentsharev1.ShareReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
		return
	}
	if err := c.svc.Share(ctx.Request.Context(), userID, req.ContentType, req.ContentID, req.RecipientUserID); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, contentsharev1.StatusRes{Status: "ok"})
}

func (c *Controller) Unshare(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	contentType := ctx.Param("content_type")
	contentID, err := parseUintParam(ctx, "content_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid content_id"})
		return
	}
	recipientUserID, err := parseUintParam(ctx, "recipient_user_id")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid recipient_user_id"})
		return
	}
	if err := c.svc.Unshare(ctx.Request.Context(), userID, contentType, contentID, recipientUserID); err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, contentsharev1.StatusRes{Status: "ok"})
}

func (c *Controller) ListRecipients(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	contentType := ctx.Query("content_type")
	contentID, err := strconv.ParseUint(ctx.Query("content_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid content_id"})
		return
	}
	res, err := c.svc.ListRecipients(ctx.Request.Context(), userID, contentType, uint(contentID))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListStudents(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.svc.ListShareableStudents(ctx.Request.Context(), userID)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *Controller) ListMine(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, swagger.ErrBody{Error: "unauthorized"})
		return
	}
	res, err := c.svc.ListSharedWithMe(ctx.Request.Context(), userID, ctx.Query("content_type"))
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func parseUintParam(ctx *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(ctx.Param(name), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWorkoutNotFound),
		errors.Is(err, service.ErrRoutineNotFound),
		errors.Is(err, service.ErrMealPlanNotFound):
		ctx.JSON(http.StatusNotFound, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrContentShareInvalidType):
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrContentShareForbidden):
		ctx.JSON(http.StatusForbidden, swagger.ErrBody{Error: err.Error()})
	case errors.Is(err, service.ErrContentShareNotTrainer),
		errors.Is(err, service.ErrContentShareNotActiveClient):
		ctx.JSON(http.StatusUnprocessableEntity, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
	}
}
