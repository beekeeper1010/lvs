package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/beekeeper1010/lvs2/global"
	"github.com/beekeeper1010/lvs2/initialize"

	"github.com/gin-gonic/gin"
	"github.com/google/gops/agent"
)

var (
	buildTime = ""
	commitId  = ""
	version   = "v1.0.0"
)

func debug() {
	if !global.Cfg.Gops.Enable {
		log.Println("gops is disabled")
		return
	}
	opts := agent.Options{
		Addr:                   fmt.Sprintf(":%d", global.Cfg.Gops.Port),
		ShutdownCleanup:        true,
		ReuseSocketAddrAndPort: true,
	}
	if err := agent.Listen(opts); err != nil {
		log.Fatal(err)
	}
	log.Println("gops is listening on", opts.Addr)
}

func printLogo() {
	log.Println("  @..@  ")
	log.Println(" (\\--/)")
	log.Println("(.>__<.)")
	log.Println("^^^  ^^^")
}

var (
	printVersion = flag.Bool("version", false, "打印版本号")
	configFile   = flag.String("config", "config.json", "配置文件路径")
	logFile      = flag.String("log", "lvs2.log", "日志文件路径")
)

func main() {
	flag.Parse()
	if *printVersion {
		log.Println(version)
		return
	}
	global.ArgCtx = &global.ArgContext{
		ConfigFile: *configFile,
		LogFile:    *logFile,
	}
	if buildTime != "" {
		log.Println("buildTime:", buildTime)
	}
	if commitId != "" {
		log.Println("commitId:", commitId)
	}
	log.Println("version:", version)
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()
	initialize.Initialize(g)
	debug()
	printLogo()
	log.Fatal(g.Run(fmt.Sprintf(":%d", global.Cfg.Port)))
}
