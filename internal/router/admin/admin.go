package admin

import (
	articlectl "trongcon-api/internal/controller/article"
	categoryctl "trongcon-api/internal/controller/category"
	uploadctl "trongcon-api/internal/controller/upload"
	userctl "trongcon-api/internal/controller/user"
	adminarticle "trongcon-api/internal/router/admin/article"
	admincategory "trongcon-api/internal/router/admin/category"
	adminupload "trongcon-api/internal/router/admin/upload"
	adminuser "trongcon-api/internal/router/admin/user"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	User     *userctl.Controller
	Category *categoryctl.Controller
	Article  *articlectl.Controller
	Upload   *uploadctl.Controller
}

func Register(r *gin.RouterGroup, c Controllers) {
	adminuser.Register(r, c.User)
	admincategory.Register(r, c.Category)
	adminarticle.Register(r, c.Article)
	adminupload.Register(r, c.Upload)
}
