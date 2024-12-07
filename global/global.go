package global

import (
	"github.com/beekeeper1010/lvs2/config"
	"github.com/beekeeper1010/lvs2/model"
	"gorm.io/gorm"
)

const (
	VERSION = "v1.1.1"
	X_TOKEN = "x-authorization"
)

var (
	Config        config.Config
	DB            *gorm.DB
	Mp4FilesCache []model.Mp4File
)
