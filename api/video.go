package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Mp4File struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

var (
	re       = regexp.MustCompile(`bytes=(\d*)-(\d*)`)
	mp4Files []Mp4File
)

func GetMp4Files(dirs ...string) error {
	id := 0
	for _, dir := range dirs {
		log.Println("walking", dir)
		err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.ToLower(filepath.Ext(info.Name())) == ".mp4" {
				log.Println("[", id, "]", path)
				mp4Files = append(mp4Files, Mp4File{
					Id:   id,
					Name: info.Name(),
					Path: path,
				})
				id++
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetVideo(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		ResponseData(c, mp4Files)
	} else {
		index, err := strconv.Atoi(id)
		if err != nil {
			ResponseError(c, err)
			return
		}
		if index < 0 || index >= len(mp4Files) {
			ResponseError(c, errors.New("invalid id"))
			return
		}
		_range := c.Request.Header.Get("Range")
		if _range == "" {
			http.ServeFile(c.Writer, c.Request, mp4Files[index].Path)
			return
		}
		file, _ := os.Open(mp4Files[index].Path)
		defer file.Close()
		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()
		var start, end int64
		subs := re.FindStringSubmatch(_range)
		if subs[1] != "" {
			start, _ = strconv.ParseInt(subs[1], 10, 64)
		}
		if subs[2] != "" {
			end, _ = strconv.ParseInt(subs[2], 10, 64)
		} else {
			end = start + 2*1024*1024
		}
		c.Status(http.StatusPartialContent)
		if end >= fileSize {
			end = fileSize - 1
			// c.Status(200)
		}
		contentLength := end - start + 1
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		c.Header("Content-Type", "video/mp4")

		file.Seek(start, io.SeekStart)
		io.CopyN(c.Writer, file, contentLength)
	}
}
