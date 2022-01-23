package handler

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRouter(client *transmissionrpc.Client, router fiber.Router) {
	setupRTorrentMetrics(router)
	setupTransmissionMetrics(client, router)
	setupQBitMetrics(router)

	router.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}
