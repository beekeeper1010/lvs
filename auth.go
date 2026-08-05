package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 24 * time.Hour

type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	PwdVer   int64  `json:"pwd_ver"`
	jwt.RegisteredClaims
}

func generateToken(uid int64, username, nickname, role string, pwdVer int64) (string, error) {
	claims := Claims{
		UserID:   uid,
		Username: username,
		Nickname: nickname,
		Role:     role,
		PwdVer:   pwdVer,
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
			// 认证失败以 HTTP 401 标识，与业务错误（HTTP 200）区分
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录", "data": nil})
			return
		}
		claims, err := parseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "登录已过期, 请重新登录", "data": nil})
			return
		}
		// 密码版本校验：改密后旧 token 立即失效，须重新登录
		var pwdVer int64
		if err := db.QueryRow(`SELECT pwd_ver FROM users WHERE id = ?`, claims.UserID).Scan(&pwdVer); err != nil || pwdVer != claims.PwdVer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "登录已失效, 请重新登录", "data": nil})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
