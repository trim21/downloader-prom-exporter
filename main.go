package main

import (
	// This would automatically load/inject environment variables from a .env file
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"

	"app/web"
)

func main() {
	if err := web.Start(); err != nil {
		log.Fatal().Msg("failed to start HTTP server")
	}
}
