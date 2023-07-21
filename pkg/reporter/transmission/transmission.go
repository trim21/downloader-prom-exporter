package transmission

import (
	"context"
	"os"
	"time"

	"github.com/hekmon/transmissionrpc/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"app/pkg/utils"
)

var torrentFields = []string{"hashString", "status", "name", "labels", "uploadedEver", "downloadedEver"}

func SetupMetrics() error {
	entryPoint, found := os.LookupEnv("TRANSMISSION_API_ENTRYPOINT")
	if !found {
		log.Info().Msg("env TRANSMISSION_API_ENTRYPOINT not set, transmission exporter disabled")
		return nil
	}

	client, err := newClient(entryPoint)
	if err != nil {
		return err
	}

	log.Info().Msg("enable transmission reporter")

	c := transmissionExporter{client: client}
	prometheus.MustRegister(c)

	return nil
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
		log.Err(err).Msg("failed to fetch transmission server stats")
		return
	}

	m <- utils.Gauge("transmission_download_session_bytes", nil, float64(h.CurrentStats.DownloadedBytes), "session download bytes")
	m <- utils.Gauge("transmission_upload_session_bytes", nil, float64(h.CurrentStats.UploadedBytes), "session upload bytes")
	m <- utils.Gauge("transmission_download_total_bytes", nil, float64(h.CumulativeStats.DownloadedBytes), "client download bytes")
	m <- utils.Gauge("transmission_upload_total_bytes", nil, float64(h.CumulativeStats.UploadedBytes), "client upload bytes")

	torrents, err := r.client.TorrentGet(ctx, torrentFields, nil)
	if err != nil {
		log.Err(err).Msg("failed to fetch transmission torrents")
		return
	}

	for _, torrent := range torrents {
		label := prometheus.Labels{"hash": *torrent.HashString}
		m <- utils.Gauge("transmission_torrent_upload_bytes", label, float64(*torrent.UploadedEver), "")
		m <- utils.Gauge("transmission_torrent_download_bytes", label, float64(*torrent.DownloadedEver), "")
	}
}
