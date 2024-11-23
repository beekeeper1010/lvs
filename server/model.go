package server

import "gorm.io/gorm"

type Mp4File struct {
	gorm.Model
	Name      string `json:"name" gorm:"not null;comment:文件名"`
	Path      string `json:"-" gorm:"not null;comment:文件路径"`
	Size      int64  `json:"size" gorm:"not null;comment:文件大小(字节)"`
	Duration  int    `json:"duration" gorm:"not null;comment:播放时长(秒)"`
	Thumbnail string `json:"thumbnail" gorm:"not null;comment:缩略图"`
}
