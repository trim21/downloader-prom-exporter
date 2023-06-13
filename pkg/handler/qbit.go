package handler

import (
	"net/url"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"app/pkg/logger"
	"app/pkg/qbittorrent"
	"app/pkg/utils"
)

func setupQBitMetrics() prometheus.Collector {
	entryPoint, found := os.LookupEnv("QBIT_API_ENTRYPOINT")
	if !found {
		return nil
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		logger.Fatal("can't parse QBIT_API_ENTRYPOINT", zap.String("value", entryPoint))
	}

	rpc, err := qbittorrent.NewClient(u)
	if err != nil {
		logger.WithE(err).Fatal("failed to create qbittorrent rpc client")
	}

	return qBittorrentExporter{client: rpc}
}

type qBittorrentExporter struct {
	client *qbittorrent.Client
}

func (r qBittorrentExporter) Describe(c chan<- *prometheus.Desc) {
}

func (r qBittorrentExporter) Collect(m chan<- prometheus.Metric) {
	t, err := r.client.Transfer()
	if err != nil {
		logger.Error("failed to get qbittorrent transfer info", zap.Error(err))
		return
	}

	m <- utils.Gauge("qbittorrent_up_info_data_bytes", nil, float64(t.UpInfoData))
	m <- utils.Gauge("qbittorrent_dl_info_data_bytes", nil, float64(t.DlInfoData))
	m <- utils.Gauge("qbittorrent_dht_nodes", nil, float64(t.DhtNodes))

	d, err := r.client.MainData()
	if err != nil {
		logger.Error("failed to get qbittorrent main info", zap.Error(err))
		return
	}

	s := d.ServerState

	m <- utils.Gauge("qbittorrent_total_buffers_size", nil, float64(s.TotalBuffersSize))
	m <- utils.Gauge("qbittorrent_upload_total_bytes", nil, float64(s.AllTimeUl))
	m <- utils.Gauge("qbittorrent_download_total_bytes", nil, float64(s.AllTimeDl))
	m <- utils.Gauge("qbittorrent_queued_io_jobs", nil, float64(s.QueuedIoJobs))
	m <- utils.Gauge("qbittorrent_average_queue_time_ms", nil, float64(s.AverageTimeQueue))

	for hash, t := range d.Torrents {
		labels := prometheus.Labels{"category": t.Category, "hash": hash}
		m <- utils.Gauge("qbittorrent_torrent_download_bytes", labels, float64(t.Downloaded))
		m <- utils.Gauge("qbittorrent_torrent_upload_bytes", labels, float64(t.Uploaded))
	}
}
