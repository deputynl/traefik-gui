package main

import (
	"embed"
	"log"

	"traefik-gui/internal/api"
	"traefik-gui/internal/config"
)

//go:embed web/dist
var webDist embed.FS

func main() {
	cfg := config.Load()
	srv := api.New(cfg, webDist)
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
