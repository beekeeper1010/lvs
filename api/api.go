package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/beekeeper1010/lvs2/global"
	"github.com/beekeeper1010/lvs2/model"
	"github.com/beekeeper1010/lvs2/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	regRange = regexp.MustCompile(`bytes=(\d*)-(\d*)`)
	errLogin = errors.New("username or password error")
)

func HandleLogin(c *gin.Context) {
	c.SetCookie("x-authorization", "", -1, "/", "", false, false)
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseLoginError(c, err)
		return
	}
	var user model.User
	if err := global.DB.First(&user, "username = ?", req.Username).Error; err != nil {
		log.Println(err)
		utils.ResponseLoginError(c, errLogin)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Println(err)
		utils.ResponseLoginError(c, errLogin)
		return
	}
	claims := model.Claims{
		Username: user.Username,
		Nickname: user.Nickname,
		Admin:    user.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "lvs2",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(global.Config.Jwt.SecretKey))
	if err != nil {
		utils.ResponseLoginError(c, err)
		return
	}
	c.SetCookie("x-authorization", tokenStr, int(7*24*time.Hour.Seconds()), "/", "", false, false)
	utils.ResponseData(c, loginResponse{
		Username:  req.Username,
		Token:     tokenStr,
		ExpiresAt: claims.ExpiresAt.Time,
	})
}

func HandleLogout(c *gin.Context) {
	utils.ResponseOk(c)
}

func HandleGetMp4List(c *gin.Context) {
	utils.ResponseData(c, global.Mp4FilesCache)
}

func HandleGetMp4File(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	if id < 1 || id > len(global.Mp4FilesCache) {
		utils.ResponseError(c, errors.New("id out of range"))
		return
	}
	sendFile(c, global.Mp4FilesCache[id-1])
}

func HandleGetMp4Total(c *gin.Context) {
	utils.ResponseData(c, len(global.Mp4FilesCache))
}

func HandleNoRoute(c *gin.Context) {
	utils.ResponseHTML(c, "index.html", nil)
}

func genToken() (string, error) {
	return "", nil
}

func sendFile(c *gin.Context, mp4File model.Mp4File) {
	var (
		err        error
		start, end int64
	)
	_range := c.Request.Header.Get("Range")
	if _range == "" {
		http.ServeFile(c.Writer, c.Request, mp4File.Path)
		return
	}
	matches := regRange.FindStringSubmatch(_range)
	if matches[1] != "" {
		start, err = strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			utils.ResponseError(c, err)
			return
		}
	}
	if matches[2] != "" {
		end, err = strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			utils.ResponseError(c, err)
			return
		}
	} else {
		end = start + 2<<20
	}
	if end >= mp4File.Size {
		end = mp4File.Size - 1
	}
	contentLen := end - start + 1
	file, err := os.Open(mp4File.Path)
	if err != nil {
		utils.ResponseError(c, err)
		return
	}
	defer file.Close()
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		utils.ResponseError(c, err)
		return
	}
	c.Status(http.StatusPartialContent)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, mp4File.Size))
	c.Header("Content-Length", strconv.FormatInt(contentLen, 10))
	c.Header("Content-Type", "video/mp4")
	io.CopyN(c.Writer, file, contentLen)
}
