package server

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	DB            *gorm.DB
	Mp4FilesCache []Mp4File
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
	DB = db
	return err
}

func InitializeTable() error {
	return DB.AutoMigrate(&Mp4File{}, &User{})
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
	return DB.Find(&Mp4FilesCache).Error
}

func initializeBase(dbfile, logfile string) {
	initializeLog(logfile)
	log.Println("initializeBase...")
	if err := InitializeDbAndTable(dbfile); err != nil {
		log.Fatal(err)
	}
	if err := initializeCache(); err != nil {
		log.Fatal(err)
	}
}

func initializeRouter(g *gin.Engine) {
	log.Println("initializeRouter...")
	api := g.Group("/api")
	{
		api.POST("/login", doLogin)
		api.POST("/logout", doLogout)
		api.GET("/mp4/list", doGetMp4List)
		api.GET("/mp4/total", doGetMp4Total)
		api.GET("/mp4/:id", doGetMp4File)
	}
	g.NoRoute(doNoRoute)
	for _, route := range g.Routes() {
		log.Printf("[%-4s] %s", route.Method, route.Path)
	}
}
