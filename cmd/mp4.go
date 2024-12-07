package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beekeeper1010/lvs2/global"
	"github.com/beekeeper1010/lvs2/initialize"
	"github.com/beekeeper1010/lvs2/model"
	"github.com/olekukonko/tablewriter"

	"github.com/spf13/cobra"
)

var mp4Cmd = &cobra.Command{
	Use:   "mp4",
	Short: "Mp4 management",
	Long:  "Mp4 management",
}

var mp4ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan mp4 files to generate sqlite db file",
	Long:  "Scan mp4 files to generate sqlite db file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := exec.Command("ffmpeg", "-version").Run(); err != nil {
			fmt.Println("ffmpeg not found, please install")
			return
		}
		dirs, _ := cmd.Flags().GetStringArray("dir")
		filter, _ := cmd.Flags().GetInt("filter")
		height, _ := cmd.Flags().GetInt("height")
		dbfile, _ := cmd.Flags().GetString("db")
		if err := scanMp4Files(dirs, filter, max(height, 100), dbfile); err != nil {
			fmt.Println(err)
		}
	},
}

var mp4ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all mp4 files",
	Long:  "List all mp4 files",
	Run: func(cmd *cobra.Command, args []string) {
		dbfile, _ := cmd.Flags().GetString("db")
		if err := listMp4Files(dbfile); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	/* mp4ScanCmd */
	mp4ScanCmd.Flags().String("db", "lvs2.db", "sqlite db file for result")
	mp4ScanCmd.Flags().StringArrayP("dir", "d", nil, "dir to scan")
	mp4ScanCmd.MarkFlagRequired("dir")
	mp4ScanCmd.Flags().IntP("filter", "f", 60, "skip mp4 files which duration is less than this value(seconds)")
	mp4ScanCmd.Flags().Int("height", 100, "height of thumbnail, min 100")
	/* mp4ListCmd */
	mp4ListCmd.Flags().String("db", "lvs2.db", "sqlite db file")
	/* mp4Cmd */
	mp4Cmd.AddCommand(mp4ScanCmd, mp4ListCmd)
	rootCmd.AddCommand(mp4Cmd)
}

func scanMp4Files(dirs []string, filter, height int, dbfile string) error {
	mp4Files := make([]*model.Mp4File, 0, 1000)
	for _, dir := range dirs {
		fmt.Println("scanning", dir)
		filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				fmt.Println(err)
				return filepath.SkipDir
			}
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".mp4" {
				return nil
			}
			fileInfo, err := os.Stat(path)
			if err != nil {
				fmt.Println(err)
				return filepath.SkipDir
			}
			duration, err := getDuration(path)
			if err != nil {
				fmt.Println(err)
				return filepath.SkipDir
			}
			if duration < filter {
				return nil
			}
			thumbnail, err := getThumbnailBase64(path, duration>>1, height)
			if err != nil {
				fmt.Println(err)
				return filepath.SkipDir
			}
			fmt.Printf("found %s, duration=%ds\n", path, duration)
			mp4Files = append(mp4Files, &model.Mp4File{
				Name:      entry.Name(),
				Path:      path,
				Size:      fileInfo.Size(),
				Duration:  duration,
				Thumbnail: thumbnail,
			})
			return nil
		})
	}
	if len(mp4Files) == 0 {
		fmt.Println("no mp4 files found")
		return nil
	}
	if err := initialize.InitializeDb(dbfile); err != nil {
		return err
	}
	global.DB.Migrator().DropTable(&model.Mp4File{})
	if err := initialize.InitializeTable(); err != nil {
		return err
	}
	result := global.DB.Create(mp4Files)
	if result.Error == nil {
		fmt.Println("inserted", result.RowsAffected, "record(s) to", dbfile)
	}
	return result.Error
}

func getDuration(path string) (int, error) {
	command := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	out, err := command.Output()
	if err != nil {
		return 0.0, err
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return int(duration), err
}

func getThumbnailBase64(path string, offset, height int) (string, error) {
	tmpPng := filepath.Join(os.TempDir(), "tmp.png")
	command := exec.Command("ffmpeg", "-v", "error", "-ss", strconv.Itoa(offset), "-i", path, "-vframes", "1", "-vf", fmt.Sprintf("scale=-1:%d", height), "-y", tmpPng)
	if _, err := command.Output(); err != nil {
		return "", err
	}
	if _, err := os.Stat(tmpPng); err != nil {
		return "", err
	}
	data, err := os.ReadFile(tmpPng)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
}

func listMp4Files(dbfile string) error {
	if err := initialize.InitializeDbAndTable(dbfile); err != nil {
		return err
	}
	var mp4Files []model.Mp4File
	if err := global.DB.Select("name", "path", "size", "duration").Find(&mp4Files).Error; err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No.", "Name", "Path", "Size/MB", "Duration/sec"})
	for i, mp4File := range mp4Files {
		table.Append([]string{
			strconv.Itoa(i + 1),
			mp4File.Name,
			mp4File.Path,
			strconv.Itoa(int(mp4File.Size >> 20)),
			strconv.Itoa(mp4File.Duration),
		})
	}
	table.Render()
	return nil
}
