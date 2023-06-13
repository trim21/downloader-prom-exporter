package handler

import (
	"net/http"
	"net/url"
	"os"

	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"gopkg.in/scgi.v0"

	"app/pkg/logger"
	rt "app/pkg/rtorrent"
)

func setupRTorrentMetrics() prometheus.Collector {
	entryPoint, found := os.LookupEnv("RTORRENT_API_ENTRYPOINT")
	if !found {
		logger.Info("env RTORRENT_API_ENTRYPOINT not set, rtorrent exporter disabled")
		return nil
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		logger.Fatal("can't parse RTORRENT_API_ENTRYPOINT", zap.String("value", entryPoint))
	}

	logger.Info("rtorrent exporter enabled")

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

func (r rTorrentExporter) Collect(metrics chan<- prometheus.Metric) {
	v, err := rt.GetGlobalData(r.rt)
	if err != nil {
	}

	t := prometheus.NewGauge(prometheus.GaugeOpts{Name: "rtorrent_upload_total_bytes", ConstLabels: prometheus.Labels{"hostname": v.Hostname}})
	t.Set(float64(v.UpTotal))
	metrics <- t

	t = prometheus.NewGauge(prometheus.GaugeOpts{Name: "rtorrent_download_total_bytes", ConstLabels: prometheus.Labels{"hostname": v.Hostname}})
	t.Set(float64(v.DownTotal))
	metrics <- t

	for _, t := range v.Torrents {
		c := prometheus.NewGauge(prometheus.GaugeOpts{Name: "rtorrent_torrent_download_bytes", ConstLabels: prometheus.Labels{"hash": t.Hash}})
		c.Set(float64(t.DownloadTotal))
		metrics <- c

		c = prometheus.NewGauge(prometheus.GaugeOpts{Name: "rtorrent_torrent_upload_bytes", ConstLabels: prometheus.Labels{"hash": t.Hash}})
		c.Set(float64(t.UploadTotal))
		metrics <- c
	}
}
