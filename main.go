package main

import (
	// This would automatically load/inject environment variables from a .env file
	_ "github.com/joho/godotenv/autoload"

	"app/pkg/logger"
	"app/web"
)

func main() {
	var w, err = web.New()
	if err != nil {
		logger.WithE(err).Error("failed to start")
	}

	if err := w.Start(); err != nil {
		logger.WithE(err).Fatal("failed to start HTTP server")
	}
}
