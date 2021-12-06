package main

import (
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
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
		logrus.Infoln("test")

		return ctx.SendString("test")
	})

	handler.SetupRouter(app)
	logrus.Infoln("start serer")

	runtime.SetBlockProfileRate(1)

	go func() {
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()

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
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339,
		DisableHTMLEscape: true,
		DataKey:           "data",
	})

	logrus.SetLevel(logrus.InfoLevel)

	envVal := os.Getenv("production")
	if envVal != "" {
		prod, err := strconv.ParseBool(envVal)
		if err != nil {
			logrus.Errorf(`can't parse "%s" as bool`, envVal)
		} else if !prod {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.SetFormatter(&logrus.TextFormatter{
				TimestampFormat:  "15:04:05 Z07:00",
				ForceColors:      true,
				ForceQuote:       true,
				FullTimestamp:    true,
				SortingFunc:      nil,
				PadLevelText:     true,
				FieldMap:         nil,
				CallerPrettyfier: nil,
			})
			logrus.Debugln("set log level to debug")
		}
	}

	rand.Seed(time.Now().UnixNano())

	log.Fatalln(startHTTP())
}
