package server

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(addr, dbfile, logfile string) {
	initializeBase(dbfile, logfile)
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()
	g.LoadHTMLFiles("index.html")
	g.Use(cors.Default())
	initializeRouter(g)
	log.Println("server running on", addr)
	log.Fatal(g.Run(addr))
}
