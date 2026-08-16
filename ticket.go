package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// playTicketTTL 播放票据有效期（短时，过期后须重新登录签发）
const playTicketTTL = 10 * time.Minute

// playStepBytes 单次 Range 响应最大字节数（滴水式拉取上限）
const playStepBytes = 4 * 1024 * 1024

// PlayTicketClaims 播放票据声明：绑定用户、视频与客户端 IP
type PlayTicketClaims struct {
	UserID  int64  `json:"uid"`
	VideoID int64  `json:"vid"`
	IP      string `json:"ip"`
	jwt.RegisteredClaims
}

// generatePlayTicket 为指定用户/视频/IP 签发短时播放票据
func generatePlayTicket(uid, vid int64, ip string) (string, error) {
	claims := PlayTicketClaims{
		UserID:  uid,
		VideoID: vid,
		IP:      ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        randomHex(8),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(playTicketTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

// parsePlayTicket 校验并解析播放票据（签名 + 有效期）
func parsePlayTicket(ticket string) (*PlayTicketClaims, error) {
	claims := &PlayTicketClaims{}
	_, err := jwt.ParseWithClaims(ticket, claims, func(t *jwt.Token) (any, error) {
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
