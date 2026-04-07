package upload

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"trongcon-api/api/swagger"
	uploadv1 "trongcon-api/api/upload/v1"
	"trongcon-api/internal/service"

	"github.com/gin-gonic/gin"
)

const maxUploadBytes = 15 << 20
const maxVideoUploadBytes = 100 << 20

var allowedFolders = map[string]struct{}{
	"categories": {},
	"articles":   {},
	"videos":     {},
	"common":     {},
}

// strongbody-api lưu key theo format: public/images/<basePath>/<filename>
// Trong code FE hiện đang gọi folder = categories | articles | videos.
// Ta map lại sang basePath mà strongbody dùng (public/categories, public/posts, ...).
func mapFolderToStrongbodyBasePath(folder string) string {
	switch folder {
	case "articles":
		return "public/posts"
	case "categories":
		return "public/categories"
	case "videos":
		// video giống thumbnail nên dùng chung folder với posts
		return "public/posts"
	case "common":
		return "public/common"
	default:
		return folder
	}
}

func maxBytesForFolder(folder string) int64 {
	if folder == "videos" {
		return maxVideoUploadBytes
	}
	return maxUploadBytes
}

type Controller struct {
	svc *service.UploadService
}

func NewController(svc *service.UploadService) *Controller {
	return &Controller{svc: svc}
}

// Upload POST multipart field "file"; query folder = categories | articles | videos | common (default common).
func (c *Controller) Upload(ctx *gin.Context) {
	folder := strings.TrimSpace(ctx.DefaultQuery("folder", "common"))
	if _, ok := allowedFolders[folder]; !ok {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "invalid folder: use categories, articles, videos, or common"})
		return
	}
	maxB := maxBytesForFolder(folder)
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "missing multipart field file"})
		return
	}
	if file.Size > maxB {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "file too large"})
		return
	}
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
		return
	}
	defer src.Close()

	body, err := io.ReadAll(io.LimitReader(src, maxB+1))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
		return
	}
	if int64(len(body)) > maxB {
		ctx.JSON(http.StatusBadRequest, swagger.ErrBody{Error: "file too large"})
		return
	}

	ct := file.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	filename := filepath.Base(file.Filename)
	if filename == "" || filename == "." {
		filename = "upload.bin"
	}

	strongbodyBasePath := mapFolderToStrongbodyBasePath(folder)
	url, err := c.svc.Upload(ctx.Request.Context(), strongbodyBasePath, filename, bytes.NewReader(body), ct)
	if err != nil {
		writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, uploadv1.UploadRes{URL: url})
}

func writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrS3NotConfigured):
		ctx.JSON(http.StatusServiceUnavailable, swagger.ErrBody{Error: err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, swagger.ErrBody{Error: err.Error()})
	}
}
