package utils

import (
	"github.com/gin-gonic/gin"
)

func ResponseOk(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"data":    nil,
		"message": "ok",
	})
}

func ResponseData(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"code":    200,
		"data":    data,
		"message": "ok",
	})
}

func ResponseHTML(c *gin.Context, template string, data any) {
	c.HTML(200, template, data)
}

func ResponseError(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"code":    500,
		"data":    nil,
		"message": err.Error(),
	})
}

func ResponseLoginError(c *gin.Context, err error) {
	c.JSON(401, gin.H{
		"code":    401,
		"data":    gin.H{"isLogin": true},
		"message": err.Error(),
	})
}

func ResponseAuthError(c *gin.Context, err error) {
	c.JSON(401, gin.H{
		"code":    401,
		"data":    gin.H{"isLogin": false},
		"message": err.Error(),
	})
}
