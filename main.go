package main

import (
	"go.uber.org/fx"

	"app/cron"
	"app/pkg/logger"
	"app/pkg/transmission"
	"app/web"
)

func main() {
	var c *cron.C
	var w web.S
	err := fx.New(
		fx.Provide(transmission.New),
		fx.Provide(cron.New),
		fx.Provide(web.New),
		fx.Populate(&c),
		fx.Populate(&w),
		fx.NopLogger,
	).Err()
	if err != nil {
		logger.WithE(err).Fatal("dependency inject")
	}

	go c.Run()
	go func() {
		if err := w.Start(); err != nil {
			logger.WithE(err).Fatal("failed to start HTTP server")
		}
	}()

	select {}
}
