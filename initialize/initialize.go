package initialize

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/beekeeper1010/lvs2/global"
	"github.com/beekeeper1010/lvs2/model/example"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	return global.Db.AutoMigrate(
		example.Example{},
	)
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
	initializeRouter(g)
}
