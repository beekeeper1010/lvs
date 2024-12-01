package global

import (
	"github.com/beekeeper1010/lvs2/config"
	"github.com/beekeeper1010/lvs2/model"
	"gorm.io/gorm"
)

var (
	Config        config.Config
	DB            *gorm.DB
	Mp4FilesCache []model.Mp4File
)
