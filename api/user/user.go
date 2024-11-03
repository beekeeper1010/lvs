package user

import (
	service "github.com/beekeeper1010/lvs2/service/user"
	"github.com/beekeeper1010/lvs2/utils"

	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context) {
	utils.ResponseData(c, service.GetUserInfo())
}
