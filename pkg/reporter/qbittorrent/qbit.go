package qbittorrent

import (
	"net/url"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"app/pkg/utils"
)

func SetupMetrics() error {
	entryPoint, found := os.LookupEnv("QBIT_API_ENTRYPOINT")
	if !found {
		log.Info().Msg("env QBIT_API_ENTRYPOINT not set, qbittorrent exporter disabled")
		return nil
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		log.Fatal().Str("value", entryPoint).Msg("can't parse QBIT_API_ENTRYPOINT")
		return err
	}

	rpc, err := newClient(u)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create qbittorrent rpc client")
		return err
	}

	log.Info().Msg("enable qbittorrent reporter")

	c := qBittorrentExporter{client: rpc}
	prometheus.MustRegister(c)

	return nil
}

type qBittorrentExporter struct {
	client *Client
}

func (r qBittorrentExporter) Describe(c chan<- *prometheus.Desc) {
}

func (r qBittorrentExporter) Collect(m chan<- prometheus.Metric) {
	t, err := r.client.Transfer()
	if err != nil {
		log.Error().Err(err).Msg("failed to get qbittorrent transfer info")
		return
	}

	m <- utils.Count("qbittorrent_up_info_data_bytes", nil, float64(t.UpInfoData))
	m <- utils.Count("qbittorrent_dl_info_data_bytes", nil, float64(t.DlInfoData))
	m <- utils.Gauge("qbittorrent_dht_nodes", nil, float64(t.DhtNodes))

	d, err := r.client.MainData()
	if err != nil {
		log.Error().Err(err).Msg("failed to get qbittorrent main info")
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
