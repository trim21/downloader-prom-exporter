package handler

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(router fiber.Router) error {
	err := setupTransmissionMetrics(router)
	if err != nil {
		return err
	}

	reporter := setupRTorrentMetrics()
	if reporter != nil {
		prometheus.MustRegister(reporter)
	}

	setupQBitMetrics(router)

	router.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	return nil
}
