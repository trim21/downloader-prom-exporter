package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pkg/errors"

	"app/pkg/handler"
	"app/pkg/logger"
)

func startHTTP() error {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		GETOnly:               false,
	})

	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:       _format(),
		TimeFormat:   time.RFC3339,
		TimeInterval: time.Second,
		Output:       os.Stdout,
	}))

	handler.SetupRouter(app)
	logger.Info("start serer")

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
	rand.Seed(time.Now().UnixNano())

	log.Fatalln(startHTTP())
}
