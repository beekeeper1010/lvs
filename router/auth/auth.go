package auth

import (
	api "github.com/beekeeper1010/lvs2/api/auth"

	"github.com/gin-gonic/gin"
)

func Initialize(base *gin.RouterGroup) {
	g := base.Group("/auth")
	g.POST("/login", api.Login)
	g.POST("/logout", api.Logout)
}
