package initialize

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/beekeeper1010/lvs2/middleware"
	"github.com/beekeeper1010/lvs2/router/auth"
	"github.com/beekeeper1010/lvs2/router/example"
	"github.com/beekeeper1010/lvs2/router/user"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed dist/*
	distFs embed.FS
)

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
	api := g.Group("/api", middleware.Auth)
	auth.Initialize(api)
	user.Initialize(api)
	example.Initialize(api)
	g.NoRoute(handleNoRoute(g))
	for _, route := range g.Routes() {
		log.Printf("[%-6s] %s", route.Method, route.Path)
	}
}
