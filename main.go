package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/pkg/errors"

	"app/pkg/handler"
)

func startHTTP() error {
	app := fiber.New(fiber.Config{
		StrictRouting: true,
		CaseSensitive: true,
		GETOnly:       false,
	})

	app.Use(requestid.New(requestid.Config{Generator: utils.UUIDv4}))

	app.Use(logger.New(logger.Config{
		Format:       _format(),
		TimeFormat:   time.RFC3339,
		TimeInterval: time.Second,
	}))

	handler.SetupRouter(app)

	return errors.Wrap(app.Listen(":80"), "failed to start http server")
}

func _format() string {
	c := color.New(color.FgCyan)
	c.EnableColor()

	format := strings.Join([]string{
		c.Sprint(strconv.Quote("time")) + `: "${time}"`,

		c.Sprint(strconv.Quote("requestid")) + `: "${locals:requestid}"`,

		c.Sprint(strconv.Quote("status")) + `: ${status}`,

		c.Sprint(strconv.Quote("method")) + `: "${method}"`,

		c.Sprint(strconv.Quote("latency")) + `: "${latency}"`,

		c.Sprint(strconv.Quote("path")) + `: "${path}"`,
	}, ", ")

	format = "{" + format + "}\n"

	return format
}

func main() {
	log.Fatalln(startHTTP())
}
