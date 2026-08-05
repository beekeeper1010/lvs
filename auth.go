package main

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 24 * time.Hour

type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func generateToken(uid int64, username string) (string, error) {
	claims := Claims{
		UserID:   uid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

func parseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// authMiddleware 从 Authorization 头或 query token 参数中取 token 校验
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			token = c.Query("token") // video 标签无法带 header，play/thumb 用 query 传
		}
		if token == "" {
			respond(c, 1, "未登录", nil)
			c.Abort()
			return
		}
		claims, err := parseToken(token)
		if err != nil {
			respond(c, 1, "登录已过期, 请重新登录", nil)
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
