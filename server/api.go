package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

var regRange = regexp.MustCompile(`bytes=(\d*)-(\d*)`)

func doLogin(c *gin.Context) {
	responseOk(c)
}

func doLogout(c *gin.Context) {
	responseOk(c)
}

func doGetMp4(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		responseData(c, Mp4FilesCache)
	} else {
		id, err := strconv.Atoi(id)
		if err != nil {
			responseError(c, err)
			return
		}
		if id < 1 || id > len(Mp4FilesCache) {
			responseError(c, errors.New("id out of range"))
			return
		}
		sendFile(c, Mp4FilesCache[id-1])
	}
}

func doGetMp4Total(c *gin.Context) {
	responseData(c, len(Mp4FilesCache))
}

func doNoRoute(c *gin.Context) {
	responseOk(c)
}

func sendFile(c *gin.Context, mp4File Mp4File) {
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
			responseError(c, err)
			return
		}
	}
	if matches[2] != "" {
		end, err = strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			responseError(c, err)
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
		responseError(c, err)
		return
	}
	defer file.Close()
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		responseError(c, err)
		return
	}
	c.Status(http.StatusPartialContent)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, mp4File.Size))
	c.Header("Content-Length", strconv.FormatInt(contentLen, 10))
	c.Header("Content-Type", "video/mp4")
	io.CopyN(c.Writer, file, contentLen)
}
