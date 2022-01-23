package web

import (
	"github.com/gofiber/fiber/v2"
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

	handler.SetupRouter(tr, app)
	logger.Info("start serer")

	return S{app}, nil
}
