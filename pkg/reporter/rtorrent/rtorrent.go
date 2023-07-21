package rtorrent

import (
	"net/http"
	"net/url"
	"os"

	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"gopkg.in/scgi.v0"

	"app/pkg/utils"
)

func SetupMetrics() error {
	entryPoint, found := os.LookupEnv("RTORRENT_API_ENTRYPOINT")
	if !found {
		log.Info().Msg("env RTORRENT_API_ENTRYPOINT not set, rtorrent exporter disabled")
		return nil
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		log.Fatal().Str("value", entryPoint).Msg("can't parse RTORRENT_API_ENTRYPOINT")
		return err
	}

	log.Info().Msg("enable rtorrent reporter")

	var rpc *xmlrpc.Client
	if u.Scheme == "scgi" {
		rpc = xmlrpc.NewClientWithHTTPClient(entryPoint, &http.Client{Transport: &scgi.Client{}})
	} else {
		rpc = xmlrpc.NewClient(entryPoint, true)
	}

	c := rTorrentExporter{rt: rpc}
	prometheus.MustRegister(c)

	return nil
}

type rTorrentExporter struct {
	rt *xmlrpc.Client
}

func (r rTorrentExporter) Describe(c chan<- *prometheus.Desc) {
}

func (r rTorrentExporter) Collect(m chan<- prometheus.Metric) {
	v, err := GetGlobalData(r.rt)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch rtorrent data")
		return
	}

	labels := prometheus.Labels{"hostname": v.Hostname}
	m <- utils.Count("rtorrent_upload_total_bytes", labels, float64(v.UpTotal), "")
	m <- utils.Count("rtorrent_download_total_bytes", labels, float64(v.DownTotal), "")
	m <- utils.Count("rtorrent_dht_nodes", labels, float64(v.DHTNodes), "")

	var peerCount int
	for _, torrent := range v.Torrents {
		peerCount += torrent.PeerConnecting
	}
	m <- utils.Count("rtorrent_total_peer_connections", labels, float64(peerCount), "")

	for _, t := range v.Torrents {
		labels := prometheus.Labels{"hash": t.Hash}
		m <- utils.Gauge("rtorrent_torrent_download_bytes", labels, float64(t.DownloadTotal), "")
		m <- utils.Gauge("rtorrent_torrent_upload_bytes", labels, float64(t.UploadTotal), "")
	}
}
