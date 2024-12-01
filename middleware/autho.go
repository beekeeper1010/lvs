package middleware

import (
	"errors"
	"log"

	"github.com/beekeeper1010/lvs2/global"
	"github.com/beekeeper1010/lvs2/model"
	"github.com/beekeeper1010/lvs2/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/api/login" {
			c.Next()
			return
		}
		tokenStr := c.Request.Header.Get("x-authorization")
		if tokenStr == "" {
			tokenStr, _ = c.Cookie("x-authorization")
		}
		if tokenStr == "" {
			utils.ResponseAuthError(c, errors.New("token is empty"))
			c.Abort()
			return
		}
		token, err := jwt.ParseWithClaims(tokenStr, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(global.Config.Jwt.SecretKey), nil
		})
		if err != nil {
			log.Println(err)
			utils.ResponseAuthError(c, errors.New("token is invalid"))
			c.Abort()
			return
		}
		if claims, ok := token.Claims.(*model.Claims); ok && token.Valid {
			c.Set("claims", claims)
		} else {
			utils.ResponseAuthError(c, errors.New("token is invalid"))
			c.Abort()
			return
		}
		c.Next()
	}
}
