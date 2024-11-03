package user

import (
	"github.com/gin-gonic/gin"
)

func GetUserInfo() gin.H {
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
		"merchantCode":  "TLif2btpzg079h15bk",
		"deleted":       0,
		"roleId":        "admin",
		"role": gin.H{
			"id":         "admin",
			"name":       "管理员",
			"describe":   "拥有所有权限",
			"status":     1,
			"creatorId":  "system",
			"createTime": 1714884570000,
			"deleted":    0,
			"permissions": []gin.H{
				{
					"roleId":         "admin",
					"permissionId":   "system",
					"permissionName": "系统管理",
					"actions":        `[{"action":"add","defaultCheck":false,"describe":"新增"},{"action":"query","defaultCheck":false,"describe":"查询"},{"action":"get","defaultCheck":false,"describe":"详情"},{"action":"update","defaultCheck":false,"describe":"修改"},{"action":"delete","defaultCheck":false,"describe":"删除"}]`,
					"actionEntitySet": []gin.H{
						{
							"action":       "add",
							"describe":     "新增",
							"defaultCheck": false,
						},
						{
							"action":       "query",
							"describe":     "查询",
							"defaultCheck": false,
						},
						{
							"action":       "get",
							"describe":     "详情",
							"defaultCheck": false,
						},
						{
							"action":       "update",
							"describe":     "修改",
							"defaultCheck": false,
						},
						{
							"action":       "delete",
							"describe":     "删除",
							"defaultCheck": false,
						},
					},
					"actionList": nil,
					"dataAccess": nil,
				},
			},
		},
	}
}
