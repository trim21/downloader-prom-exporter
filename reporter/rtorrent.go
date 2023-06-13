package reporter

import (
	"net/http"
	"net/url"
	"os"

	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"gopkg.in/scgi.v0"

	rt "app/pkg/rtorrent"
	"app/pkg/utils"
)

func setupRTorrentMetrics() prometheus.Collector {
	entryPoint, found := os.LookupEnv("RTORRENT_API_ENTRYPOINT")
	if !found {
		log.Info().Msg("env RTORRENT_API_ENTRYPOINT not set, rtorrent exporter disabled")
		return nil
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		log.Fatal().Str("value", entryPoint).Msg("can't parse RTORRENT_API_ENTRYPOINT")
	}

	log.Info().Msg("rtorrent exporter enabled")

	var rpc *xmlrpc.Client
	if u.Scheme == "scgi" {
		rpc = xmlrpc.NewClientWithHTTPClient(entryPoint, &http.Client{Transport: &scgi.Client{}})
	} else {
		rpc = xmlrpc.NewClient(entryPoint, true)
	}

	return rTorrentExporter{rt: rpc}
}

type rTorrentExporter struct {
	rt *xmlrpc.Client
}

func (r rTorrentExporter) Describe(c chan<- *prometheus.Desc) {
}

func (r rTorrentExporter) Collect(m chan<- prometheus.Metric) {
	v, err := rt.GetGlobalData(r.rt)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch rtorrent data")
		return
	}

	m <- utils.Gauge("rtorrent_upload_total_bytes", prometheus.Labels{"hostname": v.Hostname}, float64(v.UpTotal))

	m <- utils.Gauge("rtorrent_download_total_bytes", prometheus.Labels{"hostname": v.Hostname}, float64(v.DownTotal))

	for _, t := range v.Torrents {
		m <- utils.Gauge("rtorrent_torrent_download_bytes", prometheus.Labels{"hash": t.Hash}, float64(t.DownloadTotal))
		m <- utils.Gauge("rtorrent_torrent_upload_bytes", prometheus.Labels{"hash": t.Hash}, float64(t.UploadTotal))
	}
}
