package example

import (
	model "github.com/beekeeper1010/lvs2/model/example"
	service "github.com/beekeeper1010/lvs2/service/example"
	"github.com/beekeeper1010/lvs2/utils"

	"github.com/gin-gonic/gin"
)

func Create(c *gin.Context) {
	req := model.ExampleRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, err)
		return
	}
	if err := service.Create(req); err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseOk(c)
}

func List(c *gin.Context) {
	examples, err := service.List()
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	utils.ResponseData(c, examples)
}
