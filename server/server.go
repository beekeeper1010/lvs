package server

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func printLogo() {
	fmt.Println(`
 _     __   __  ___   ___ 
| |    \ \ / / / __| |_  )
| |__   \ V /  \__ \  / /
|____|   \_/   |___/ /___|`)
}

func Run(addr, dbfile, logfile string) {
	initializeBase(dbfile, logfile)
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()
	g.LoadHTMLFiles("index.html")
	g.Use(cors.Default())
	initializeRouter(g)
	log.Println("server running on", addr)
	printLogo()
	log.Fatal(g.Run(addr))
}
