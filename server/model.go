package server

import "gorm.io/gorm"

type Mp4File struct {
	gorm.Model
	Name      string `json:"name" gorm:"not null;comment:文件名"`
	Path      string `json:"-" gorm:"not null;comment:文件路径"`
	Duration  int    `json:"duration" gorm:"not null;comment:时长"`
	Thumbnail string `json:"thumbnail" gorm:"not null;comment:缩略图"`
}
