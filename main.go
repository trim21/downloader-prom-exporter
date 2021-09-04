package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"app/pkg/handler"
)

func startHTTP() error {
	app := fiber.New(fiber.Config{
		StrictRouting: true,
		CaseSensitive: true,
		GETOnly:       true,
	})

	handler.SetupRouter(app)

	return errors.Wrap(app.Listen(":3001"), "failed to start http server")
}

func main() {
	log.Fatalln(startHTTP())
}
