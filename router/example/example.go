package example

import (
	api "github.com/beekeeper1010/lvs2/api/example"

	"github.com/gin-gonic/gin"
)

func Initialize(base *gin.RouterGroup) {
	g := base.Group("/example")
	g.POST("/create", api.Create)
	g.GET("/list", api.List)
}
