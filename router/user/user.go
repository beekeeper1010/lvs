package user

import (
	api "github.com/beekeeper1010/lvs2/api/user"

	"github.com/gin-gonic/gin"
)

func Initialize(base *gin.RouterGroup) {
	g := base.Group("/user")
	g.GET("/info", api.GetUserInfo)
}
