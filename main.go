package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pkg/errors"

	"app/pkg/handler"
)

func startHTTP() error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		GETOnly:               false,
	})

	app.Use(logger.New(logger.Config{
		Format:       _format(),
		TimeFormat:   time.RFC3339,
		TimeInterval: time.Second,
		Output:       os.Stdout,
	}))

	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendString("test")
	})

	handler.SetupRouter(app)

	return errors.Wrap(app.Listen(":80"), "failed to start http server")
}

func _format() string {
	format := strings.Join([]string{
		strconv.Quote("time") + `: "${time}"`,

		(strconv.Quote("status")) + `: ${status}`,

		(strconv.Quote("method")) + `: "${method}"`,

		(strconv.Quote("latency")) + `: "${latency}"`,

		(strconv.Quote("path")) + `: "${path}"`,
	}, ", ")

	format = "{" + format + "}\n"

	return format
}

func main() {
	log.Fatalln(startHTTP())
}
