package server

import (
	"github.com/gin-gonic/gin"
)

func responseOk(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"data":    nil,
		"message": "ok",
	})
}

func responseData(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"code":    200,
		"data":    data,
		"message": "ok",
	})
}

func responseHTML(c *gin.Context, template string, data any) {
	c.HTML(200, template, data)
}

func responseError(c *gin.Context, err error) {
	c.JSON(500, gin.H{
		"code":    500,
		"data":    nil,
		"message": err.Error(),
	})
}
