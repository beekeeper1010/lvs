package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// scanDirectory 递归扫描目录下的 mp4 文件并入库，已存在的记录更新缩略图
func scanDirectory(dir, thumbsDir string) error {
	if err := os.MkdirAll(thumbsDir, 0o755); err != nil {
		return err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s 不是目录", dir)
	}
	count := 0
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".mp4" {
			return nil
		}
		thumb, ok := generateThumb(path, thumbsDir)
		upsertVideo(filepath.Base(path), path, thumb, ok)
		count++
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	fmt.Printf("扫描完成: 共 %d 个视频, 缩略图目录 %s\n", count, thumbsDir)
	return nil
}

// generateThumb 用 ffmpeg 提取视频第一帧为 320 宽缩略图, 返回缩略图路径
func generateThumb(videoPath, thumbsDir string) (string, bool) {
	h := md5.Sum([]byte(videoPath))
	thumbPath := filepath.Join(thumbsDir, hex.EncodeToString(h[:])+".jpg")
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", videoPath,
		"-frames:v", "1",
		"-vf", "scale=320:-2",
		"-q:v", "3",
		thumbPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("ffmpeg 提取缩略图失败 %s: %v %s", videoPath, err, out)
		_ = os.Remove(thumbPath)
		return "", false
	}
	return thumbPath, true
}

// upsertVideo 按 path 唯一键插入或更新视频记录
func upsertVideo(name, path, thumb string, thumbOK bool) bool {
	if !thumbOK {
		thumb = ""
	}
	res, err := db.Exec(`
		INSERT INTO videos(name, path, thumb_path, created_at) VALUES(?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET name=excluded.name, thumb_path=excluded.thumb_path`,
		name, path, thumb, time.Now())
	if err != nil {
		log.Printf("写入数据库失败 %s: %v", path, err)
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}
