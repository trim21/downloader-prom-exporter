package handler

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRouter(router fiber.Router) {
	setupRTorrentMetrics(router)
	setupTransmissionMetrics(router)
	setupQBitMetrics(router)
}
