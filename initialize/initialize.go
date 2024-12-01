package initialize

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/beekeeper1010/lvs2/api"
	"github.com/beekeeper1010/lvs2/global"
	"github.com/beekeeper1010/lvs2/middleware"
	"github.com/beekeeper1010/lvs2/model"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func initializeLog(logfile string) {
	w := io.MultiWriter(&lumberjack.Logger{
		Filename:   logfile,
		MaxSize:    10,
		MaxAge:     10,
		MaxBackups: 10,
		LocalTime:  true,
		Compress:   true,
	}, os.Stdout)
	log.SetOutput(w)
}

func InitializeDb(dbfile string) error {
	db, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}
	global.DB = db
	return err
}

func InitializeTable() error {
	return global.DB.AutoMigrate(
		&model.Mp4File{},
		&model.User{},
	)
}

func initializeConfig(cfgfile string) error {
	data, err := os.ReadFile(cfgfile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &global.Config)
}

func InitializeDbAndTable(dbfile string) error {
	if err := InitializeDb(dbfile); err != nil {
		return err
	}
	if err := InitializeTable(); err != nil {
		return err
	}
	return nil
}

func initializeCache() error {
	return global.DB.Find(&global.Mp4FilesCache).Error
}

func InitializeBase(dbfile, cfgfile, logfile string) {
	initializeLog(logfile)
	log.Println("initializeBase...")
	if err := initializeConfig(cfgfile); err != nil {
		log.Fatal(err)
	}
	if err := InitializeDbAndTable(dbfile); err != nil {
		log.Fatal(err)
	}
	if err := initializeCache(); err != nil {
		log.Fatal(err)
	}
}

func InitializeRouter(g *gin.Engine) {
	log.Println("initializeRouter...")
	group := g.Group("/api", middleware.JwtAuth())
	{
		group.POST("/login", api.HandleLogin)
		group.POST("/logout", api.HandleLogout)
		group.GET("/mp4/list", api.HandleGetMp4List)
		group.GET("/mp4/total", api.HandleGetMp4Total)
		group.GET("/mp4/:id", api.HandleGetMp4File)
	}
	g.NoRoute(api.HandleNoRoute)
	for _, route := range g.Routes() {
		log.Printf("[%-4s] %s", route.Method, route.Path)
	}
}
