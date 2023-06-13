package handler

import (
	"context"
	"time"

	"github.com/hekmon/transmissionrpc/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"app/pkg/logger"
	"app/pkg/transmission"
	"app/pkg/utils"
)

var torrentFields = []string{"hashString", "status", "name", "labels", "uploadedEver", "downloadedEver"}

func setupTransmissionMetrics() (prometheus.Collector, error) {
	client, err := transmission.New()
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, nil
	}

	logger.Info("enable transmission reporter")

	return transmissionExporter{client: client}, nil
}

type transmissionExporter struct {
	client *transmissionrpc.Client
}

func (r transmissionExporter) Describe(c chan<- *prometheus.Desc) {
}

func (r transmissionExporter) Collect(m chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	h, err := r.client.SessionStats(ctx)
	if err != nil {
		logger.Error("failed to fetch transmission server stats", zap.Error(err))
		return
	}

	m <- utils.Gauge("transmission_download_session_bytes", nil, float64(h.CurrentStats.DownloadedBytes))
	m <- utils.Gauge("transmission_upload_session_bytes", nil, float64(h.CurrentStats.UploadedBytes))

	torrents, err := r.client.TorrentGet(ctx, torrentFields, nil)
	if err != nil {
		logger.Error("failed to fetch transmission torrents", zap.Error(err))
		return
	}

	for _, torrent := range torrents {
		label := prometheus.Labels{"hash": *torrent.HashString}
		m <- utils.Gauge("transmission_torrent_upload_bytes", label, float64(*torrent.UploadedEver))
		m <- utils.Gauge("transmission_torrent_download_bytes", label, float64(*torrent.DownloadedEver))
	}
}
