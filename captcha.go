package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// captchaStore 验证码内存存储，过期自动清理
var captchaStore = base64Captcha.DefaultMemStore

// handleCaptcha 生成图形验证码，返回验证码 ID 与 Base64 图片（data URI）
// 点阵数字驱动：默认干扰与噪声水平
func handleCaptcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(46, 140, 4, 0.6, 30)
	captcha := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		respond(c, 1, "验证码生成失败", nil)
		return
	}
	respond(c, 0, "ok", gin.H{"captcha_id": id, "captcha_image": b64s})
}

// verifyCaptcha 校验验证码，校验后立即失效（一次性）
func verifyCaptcha(id, code string) bool {
	return captchaStore.Verify(id, code, true)
}
