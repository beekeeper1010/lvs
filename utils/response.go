package utils

import (
	"net/http"

	"github.com/beekeeper1010/lvs2/global"
	"github.com/gin-gonic/gin"
)

func ResponseOk(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"data":    nil,
		"message": "ok",
	})
}

func ResponseData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"data":    data,
		"message": "ok",
	})
}

func ResponseHTML(c *gin.Context, template string, data any) {
	c.HTML(http.StatusOK, template, data)
}

func ResponseError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    http.StatusInternalServerError,
		"data":    nil,
		"message": err.Error(),
	})
}

func ResponseAuthError(c *gin.Context, err error) {
	c.SetCookie(global.X_TOKEN, "", -1, "/", "", false, false)
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"data":    nil,
		"message": err.Error(),
	})
}
