package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"docker-dashboard/internal/api"

	"github.com/labstack/echo/v4"
)

func main() {
	staticDir := filepath.Join("web", "public")
	for {
		e := echo.New()
		api.RegisterRoutes(e)

		// Serve static files from "web/public"; SPA fallback: unknown paths serve index.html
		e.Static("/", staticDir)
		e.GET("/*", func(c echo.Context) error {
			return c.File(filepath.Join(staticDir, "index.html"))
		})

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr := ":" + port
		log.Printf("Server started at http://localhost%s", addr)
		err := e.Start(addr)
		if err != nil {
			log.Printf("Server error: %v", err)
		}
		log.Println("Restarting server in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}
