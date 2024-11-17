package api

import (
	"github.com/gin-gonic/gin"
)

func ResponseOk(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "ok",
		"result":  nil,
	})
}

func ResponseData(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "ok",
		"result":  data,
	})
}

func ResponseError(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"code":    500,
		"message": err.Error(),
		"result":  nil,
	})
}
