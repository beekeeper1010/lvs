package initialize

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/beekeeper1010/lvs2/api"
	"github.com/beekeeper1010/lvs2/global"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	//go:embed dist/*
	distFs embed.FS
)

func initializeLog() error {
	w := io.MultiWriter(&lumberjack.Logger{
		Filename:   global.ArgCtx.LogFile,
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 10,
		LocalTime:  true,
		Compress:   true,
	}, os.Stdout)
	log.SetOutput(w)
	log.Println("initializeLog...")
	return nil
}

func initializeConfig() error {
	log.Println("initializeConfig...")
	data, err := os.ReadFile(global.ArgCtx.ConfigFile)
	if err != nil {
		return err
	}
	var cfg global.Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}
	global.Cfg = &cfg
	return nil
}

func initializeMysqlDb() error {
	log.Println("initializeMysqlDb...")
	for i := 0; i < 6; i++ {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN:               global.Cfg.Mysql.DSN(),
			DefaultStringSize: 191,
		}), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger: logger.New(log.Default(), logger.Config{
				SlowThreshold:             100 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
			}),
		})
		if err == nil {
			global.Db = db
			return nil
		}
		log.Println("连接数据库失败，等待重试", err)
		if i < 5 {
			time.Sleep(10 * time.Second)
		}
	}
	return errors.New("连接数据库失败")
}

func initializeSqliteDb() error {
	log.Println("initializeSqliteDb...")
	db, err := gorm.Open(sqlite.Open(global.Cfg.Sqlite.Db), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(log.Default(), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
	global.Db = db
	return err
}

func initializeTable() error {
	log.Println("initializeTable...")
	return global.Db.AutoMigrate()
}

func handleNoRoute(g *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		staticFs, _ := fs.Sub(distFs, "dist")
		path := c.Request.URL.Path
		if path != "/" {
			f, err := staticFs.Open(path[1:])
			if err != nil {
				c.Request.URL.Path = "/"
				g.HandleContext(c)
				return
			}
			defer f.Close()
			if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
				if strings.HasSuffix(path, "css") || strings.HasSuffix(path, "js") {
					magic := make([]byte, 2)
					f.Read(magic)
					if magic[0] == 0x1f && magic[1] == 0x8b {
						c.Header("Content-Encoding", "gzip")
					}
				}
			}
		}
		http.FileServer(http.FS(staticFs)).ServeHTTP(c.Writer, c.Request)
	}
}

func initializeRouter(g *gin.Engine) {
	log.Println("initializeRouter...")
	g.GET("/api/video", api.GetVideo)
	// g.NoRoute(handleNoRoute(g))
	for _, route := range g.Routes() {
		log.Printf("[%-6s] %s", route.Method, route.Path)
	}
}

func Initialize(g *gin.Engine) {
	if err := initializeLog(); err != nil {
		log.Fatal(err)
	}
	if err := initializeConfig(); err != nil {
		log.Fatal(err)
	}
	if global.Cfg.DbType == global.DB_MYSQL {
		if err := initializeMysqlDb(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := initializeSqliteDb(); err != nil {
			log.Fatal(err)
		}
	}
	if err := initializeTable(); err != nil {
		log.Fatal(err)
	}
	if err := api.GetMp4Files(global.Cfg.Dirs...); err != nil {
		log.Fatal(err)
	}
	initializeRouter(g)
}
