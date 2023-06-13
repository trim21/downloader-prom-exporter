package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/pkg/errors"

	"app/pkg/logger"
	"app/pkg/transmission"
)

func setupTransmissionMetrics(router fiber.Router) error {
	client, err := transmission.New()
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	logger.Info("enable transmission reporter")

	var interval = 10 * time.Second
	if rawInterval, found := os.LookupEnv("TRANSMISSION_UPDATE_INTERVAL"); found {
		v, err := time.ParseDuration(rawInterval)
		if err != nil || v <= 0 {
			logger.WithE(err).Sugar().Warnf(
				"can't parse '%s' as time.Duration, use default value %s", rawInterval, interval)
		} else {
			logger.Sugar().Infof("set transmission update interval to '%s'", v)
			interval = v
		}
	}

	h := &transmissionHandler{
		torrentMux:    sync.RWMutex{},
		statusMux:     sync.RWMutex{},
		stats:         transmissionrpc.SessionStats{},
		client:        client,
		torrentFields: []string{"hashString", "status", "name", "labels", "uploadedEver", "downloadedEver"},
	}

	logger.Info("start fetching transmission torrent details")
	go runInBackground(interval, h.torrentUpdater)
	logger.Info("start fetching transmission global data")
	go runInBackground(interval, h.statsUpdater)

	router.Get("/transmission/metrics", h.fiberHandler)

	return nil
}

type transmissionHandler struct {
	torrentsErr   error
	statusErr     error
	client        *transmissionrpc.Client
	torrents      []transmissionrpc.Torrent
	torrentFields []string
	stats         transmissionrpc.SessionStats
	torrentMux    sync.RWMutex
	statusMux     sync.RWMutex
}

func (h *transmissionHandler) torrentUpdater() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	v, err := h.client.TorrentGet(ctx, h.torrentFields, nil)
	h.torrentMux.Lock()
	defer h.torrentMux.Unlock()
	if err != nil {
		h.torrentsErr = errors.Wrap(err, "failed to get torrents")
		logger.WithE(err).Error("failed to get torrents")
	} else {
		h.torrentsErr = nil
		h.torrents = v
	}
}

func (h *transmissionHandler) statsUpdater() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	v, err := h.client.SessionStats(ctx)
	h.statusMux.Lock()
	defer h.statusMux.Unlock()
	if err != nil {
		h.statusErr = errors.Wrap(err, "failed to get session stats")
		logger.WithE(err).Error("failed to get session stats")
	} else {
		h.statusErr = nil
		h.stats = v
	}
}

func (h *transmissionHandler) fiberHandler(ctx *fiber.Ctx) error {
	logger.Debug("export transmission metrics")
	h.statusMux.RLock()
	if h.statusErr != nil {
		h.statusMux.RUnlock()

		return h.statusErr
	}

	fmt.Fprintln(ctx, "# without label filter")
	fmt.Fprintf(ctx, "transmission_download_all_total %d\n", h.stats.CumulativeStats.DownloadedBytes)
	fmt.Fprintf(ctx, "transmission_upload_all_total %d\n", h.stats.CurrentStats.UploadedBytes)
	h.statusMux.RUnlock()

	h.torrentMux.RLock()
	defer h.torrentMux.RUnlock()

	if h.torrentsErr != nil {
		return h.torrentsErr
	}

	fmt.Fprintln(ctx, "\n# all torrents")
	for i := range h.torrents {
		writeTorrent(ctx, &h.torrents[i])
	}

	return nil
}

func runInBackground(interval time.Duration, f func()) {
	f()

	// wait for even time to start looping
	i := int64(interval)
	<-time.After(time.Duration(i - (time.Now().UnixNano() % i)))

	for range time.NewTicker(interval).C {
		f()
	}
}

func writeTorrent(w io.Writer, t *transmissionrpc.Torrent) {
	fmt.Fprintln(w, "\n# torrent", strconv.Quote(*t.Name))
	fmt.Fprintln(w, "# labels:", strings.Join(t.Labels, ", "))

	if len(t.Labels) == 0 {
		label := fmt.Sprintf("hash=%s, stats=%s", strconv.Quote(*t.HashString), strconv.Quote(t.Status.String()))

		fmt.Fprintf(w, "transmission_torrent_download_bytes{%s} %d\n", label, *t.DownloadedEver)
		fmt.Fprintf(w, "transmission_torrent_upload_bytes{%s} %d\n", label, *t.UploadedEver)
	} else {
		for _, label := range t.Labels {
			fmt.Fprintln(w, "# label ", label)
			label := fmt.Sprintf("label=%s, hash=%s, stats=%s",
				strconv.Quote(label), strconv.Quote(*t.HashString), strconv.Quote(t.Status.String()))

			fmt.Fprintf(w, "transmission_torrent_download_bytes{%s} %d\n", label, *t.DownloadedEver)
			fmt.Fprintf(w, "transmission_torrent_upload_bytes{%s} %d\n", label, *t.UploadedEver)
		}
	}
}
