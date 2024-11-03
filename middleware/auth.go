package middleware

import (
	"errors"

	"github.com/beekeeper1010/lvs2/utils"

	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	if c.Request.URL.Path != "/api/auth/login" && c.Request.Header.Get("Access-Token") != "4291d7da9005377ec9aec4a71ea837f" {
		utils.ResponseAuthError(c, errors.New("未授权"))
		c.Abort()
		return
	}
	c.Next()
}
