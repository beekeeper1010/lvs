package server

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Run(addr, dbfile, logfile string) {
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()
	initialize(g, dbfile, logfile)
	log.Println("server running on", addr)
	log.Fatal(g.Run(addr))
}
