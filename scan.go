package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// probeResult ffprobe 探测结果
type probeResult struct {
	hasVideo bool
	duration float64
}

// ffProbeOut ffprobe JSON 输出结构
type ffProbeOut struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		Duration  string `json:"duration"`
	} `json:"streams"`
}

// probeVideo 一次 ffprobe（JSON）获取视频流存在性与时长，减少外部进程调用。
// 时长优先取视频流时长，回退 format.duration（兼容 format 时长缺失的视频）
func probeVideo(videoPath string) probeResult {
	// ffprobe 不可用时乐观处理（假设有视频流），交给 ffmpeg 提取时判定
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return probeResult{hasVideo: true}
	}
	// 仅输出所需字段，避免全量 JSON 带来的解析开销
	out, err := exec.Command("ffprobe", "-v", "error",
		"-print_format", "json",
		"-show_entries", "stream=codec_type,duration:format=duration", videoPath).Output()
	if err != nil {
		return probeResult{}
	}
	var data ffProbeOut
	if err := json.Unmarshal(out, &data); err != nil {
		return probeResult{}
	}
	var r probeResult
	for _, s := range data.Streams {
		if s.CodecType == "video" {
			r.hasVideo = true
			if d, err := strconv.ParseFloat(s.Duration, 64); err == nil && d > 0 {
				r.duration = d
			}
		}
	}
	if r.duration <= 0 {
		if d, err := strconv.ParseFloat(data.Format.Duration, 64); err == nil && d > 0 {
			r.duration = d
		}
	}
	return r
}

// videoThumbFresh 判断视频是否已有最新缩略图（记录存在且缩略图文件新于源文件），用于跳过重复处理
func videoThumbFresh(path string) bool {
	var thumb string
	if err := db.QueryRow(`SELECT thumb_path FROM videos WHERE path = ?`, path).Scan(&thumb); err != nil || thumb == "" {
		return false
	}
	ti, err := os.Stat(thumb)
	if err != nil {
		return false
	}
	vi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return ti.ModTime().After(vi.ModTime())
}

// scanTask 扫描任务
type scanTask struct {
	path string
	name string
}

// scanResult 扫描结果
type scanResult struct {
	task     scanTask
	fresh    bool // 已是最新，跳过
	hasVideo bool
	duration float64
	thumb    string
	ok       bool
}

// scanDirectory 递归扫描目录下的 mp4 文件并入库。
// 已存在且未变更的视频跳过处理；缩略图提取并发执行以提升吞吐
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
	tasks := make([]scanTask, 0)
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ".mp4" {
			tasks = append(tasks, scanTask{path: path, name: filepath.Base(path)})
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	// 并发提取缩略图（ffprobe/ffmpeg 为 CPU/IO 密集任务）
	workers := max(1, min(runtime.NumCPU(), 8))
	results := make(chan scanResult, len(tasks))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(t scanTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			// 缩略图已是最新则跳过（避免重复调用 ffprobe/ffmpeg）
			if videoThumbFresh(t.path) {
				results <- scanResult{task: t, fresh: true}
				return
			}
			probe := probeVideo(t.path)
			if !probe.hasVideo {
				results <- scanResult{task: t, hasVideo: false}
				return
			}
			thumb, ok := generateThumb(t.path, thumbsDir, probe.duration)
			results <- scanResult{task: t, hasVideo: true, duration: probe.duration, thumb: thumb, ok: ok}
		}(t)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	// 串行入库（sqlite 单写者）
	count := 0
	for r := range results {
		switch {
		case r.fresh:
			count++
		case !r.hasVideo:
			upsertVideo(r.task.name, r.task.path, "", 0, false)
			log.Printf("无视频流, 跳过缩略图: %s", r.task.path)
			count++
		default:
			upsertVideo(r.task.name, r.task.path, r.thumb, r.duration, r.ok)
			count++
		}
	}
	fmt.Printf("扫描完成: 共 %d 个视频, 缩略图目录 %s\n", count, thumbsDir)
	return nil
}

// generateThumb 用 ffmpeg 提取视频中间时间点的帧为 320 宽缩略图, 返回缩略图路径
func generateThumb(videoPath, thumbsDir string, duration float64) (string, bool) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Printf("未找到 ffmpeg, 请先安装并加入 PATH: %v", err)
		return "", false
	}
	h := md5.Sum([]byte(videoPath))
	thumbPath := filepath.Join(thumbsDir, hex.EncodeToString(h[:])+".jpg")
	// -ss 置于 -i 前使用快速 seek，避免解码全片取帧；时长未知时回退到首帧
	base := []string{"-y", "-hide_banner", "-loglevel", "error"}
	if duration > 0 {
		base = append(base, "-ss", fmt.Sprintf("%.3f", duration/2))
	}
	base = append(base, "-i", videoPath, "-map", "0:v:0", "-frames:v", "1", "-an")
	// 优先带缩放参数, 失败时退回原尺寸（兼容旋转/奇数尺寸等特殊视频）
	attempts := [][]string{
		append(append([]string{}, base...), "-vf", "scale=320:-2", "-q:v", "3", thumbPath),
		append(append([]string{}, base...), "-q:v", "3", thumbPath),
	}
	for _, args := range attempts {
		cmd := exec.Command("ffmpeg", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("ffmpeg 提取缩略图失败 %s: %v\n%s", videoPath, err, out)
			continue
		}
		return thumbPath, true
	}
	_ = os.Remove(thumbPath)
	return "", false
}

// upsertVideo 按 path 唯一键插入或更新视频记录
func upsertVideo(name, path, thumb string, duration float64, thumbOK bool) bool {
	if !thumbOK {
		thumb = ""
	}
	res, err := db.Exec(`
		INSERT INTO videos(name, path, thumb_path, duration, created_at) VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET name=excluded.name, thumb_path=excluded.thumb_path, duration=excluded.duration`,
		name, path, thumb, duration, time.Now())
	if err != nil {
		log.Printf("写入数据库失败 %s: %v", path, err)
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}
