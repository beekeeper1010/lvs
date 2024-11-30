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

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"not null;unique;comment:用户名"`
	Nickname string `json:"nickname" gorm:"not null;comment:昵称"`
	Password string `json:"password" gorm:"not null;comment:密码"`
	Admin    bool   `json:"admin" gorm:"not null;comment:是否为管理员"`
}
