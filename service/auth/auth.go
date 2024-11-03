package auth

import (
	"errors"

	model "github.com/beekeeper1010/lvs2/model/auth"

	"github.com/gin-gonic/gin"
)

func Login(req model.LoginRequest) (gin.H, error) {
	if req.Username != "admin" || req.Password != "e6e061838856bf47e1de730719fb2609" {
		return gin.H{}, errors.New("用户名或密码错误")
	}
	return gin.H{
		"id":            "4291d7da9005377ec9aec4a71ea837f",
		"name":          "井蛙",
		"username":      "admin",
		"password":      "",
		"avatar":        "/avatar2.jpg",
		"status":        1,
		"telephone":     "",
		"lastLoginIp":   "127.0.0.1",
		"lastLoginTime": 1714884570000,
		"creatorId":     "admin",
		"createTime":    1714884570000,
		"deleted":       0,
		"roleId":        "admin",
		"lang":          "zh-CN",
		"token":         "4291d7da9005377ec9aec4a71ea837f",
	}, nil
}
