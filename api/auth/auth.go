package auth

import (
	model "github.com/beekeeper1010/lvs2/model/auth"
	service "github.com/beekeeper1010/lvs2/service/auth"
	"github.com/beekeeper1010/lvs2/utils"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseLoginError(c, err)
		return
	}
	rsp, err := service.Login(req)
	if err != nil {
		utils.ResponseLoginError(c, err)
		return
	}
	utils.ResponseData(c, rsp)
}

func Logout(c *gin.Context) {
	utils.ResponseOk(c)
}
