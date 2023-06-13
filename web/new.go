package web

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/savsgio/gotils/nocopy"

	"app/pkg/errgo"
	"app/pkg/handler"
	"app/pkg/logger"
)

type S struct {
	_   nocopy.NoCopy
	app *fiber.App
}

func (s *S) Start() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	return errgo.Wrap(s.app.Listen(fmt.Sprintf("127.0.0.1:%s", port)), "failed to start http server")
}

func New() (S, error) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		CaseSensitive:         true,
		GETOnly:               false,
	})

	err := handler.SetupRouter(app)
	if err != nil {
		return S{}, err
	}

	logger.Info("start server")

	return S{app: app}, nil
}
