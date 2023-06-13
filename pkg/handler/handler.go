package handler

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(router fiber.Router) error {
	if reporter, err := setupTransmissionMetrics(); err != nil {
		return err
	} else if reporter != nil {
		prometheus.MustRegister(reporter)
	}

	if reporter := setupRTorrentMetrics(); reporter != nil {
		prometheus.MustRegister(reporter)
	}

	if reporter := setupQBitMetrics(); reporter != nil {
		prometheus.MustRegister(reporter)
	}

	router.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	return nil
}
