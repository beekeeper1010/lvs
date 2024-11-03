package global

import (
	"fmt"

	"gorm.io/gorm"
)

type ArgContext struct {
	ConfigFile string
	LogFile    string
}

type Gops struct {
	Enable bool `json:"enable"`
	Port   int  `json:"port"`
}

type Mysql struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Db       string `json:"db"`
	Extra    string `json:"extra"`
}

func (m Mysql) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s", m.Username, m.Password, m.Host, m.Port, m.Db, m.Extra)
}

type Sqlite struct {
	Db string `json:"db"`
}

type Config struct {
	Port   int    `json:"port"`
	Gops   Gops   `json:"gops"`
	DbType string `json:"dbType"`
	Mysql  Mysql  `json:"mysql"`
	Sqlite Sqlite `json:"sqlite"`
}

var (
	ArgCtx *ArgContext
	Cfg    *Config
	Db     *gorm.DB
)
