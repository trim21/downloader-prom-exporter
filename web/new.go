package web

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hekmon/transmissionrpc/v2"

	"app/pkg/errgo"
	"app/pkg/handler"
	"app/pkg/logger"
)

type S struct {
	app *fiber.App
}

func (s S) Start() error {
	return errgo.Wrap(s.app.Listen(":80"), "failed to start http server")
}

func New(tr *transmissionrpc.Client) (S, error) {
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

	handler.SetupRouter(tr, app)
	logger.Info("start serer")

	return S{app}, nil
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
