package handler

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(router fiber.Router) {
	setupRTorrentMetrics(router)
	setupTransmissionMetrics(router)
	setupQBitMetrics(router)

	router.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

}
