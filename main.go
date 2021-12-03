package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"app/pkg/handler"
)

func startHTTP() error {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339Nano,
		DisableTimestamp:  false,
		DisableHTMLEscape: true,
		DataKey:           "data",
		PrettyPrint:       false,
		// FieldMap: logrus.FieldMap{
		// 	logrus.FieldKeyTime:  "@timestamp",
		// 	logrus.FieldKeyLevel: "@level",
		// 	logrus.FieldKeyMsg:   "@message",
		// 	logrus.FieldKeyFunc:  "@caller",
		// },
	})

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		GETOnly:               false,
	})

	app.Use(logger.New(logger.Config{
		Format:       _format(),
		TimeFormat:   time.RFC3339Nano,
		TimeInterval: time.Second,
		Output:       os.Stdout,
	}))

	app.Get("/test", func(ctx *fiber.Ctx) error {
		return ctx.SendString("test")
	})

	handler.SetupRouter(app)

	logrus.Infoln("start serer")
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
