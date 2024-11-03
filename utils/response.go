package utils

import (
	"github.com/gin-gonic/gin"
)

func ResponseOk(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"result":  nil,
	})
}

func ResponseData(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"result":  data,
	})
}

func ResponseAuthError(c *gin.Context, err error) {
	c.JSON(401, gin.H{
		"code":    1,
		"message": err.Error(),
		"result":  nil,
	})
}

func ResponseLoginError(c *gin.Context, err error) {
	c.JSON(401, gin.H{
		"code":    1,
		"message": err.Error(),
		"result":  gin.H{"isLogin": true},
	})
}

func ResponseError(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"code":    1,
		"message": err.Error(),
		"result":  nil,
	})
}
