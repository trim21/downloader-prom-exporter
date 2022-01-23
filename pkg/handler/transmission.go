package handler

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"app/pkg/logger"
	"app/pkg/utils"
)

func setupTransmissionMetrics(router fiber.Router) {
	entryPoint, found := os.LookupEnv("TRANSMISSION_API_ENTRYPOINT")
	if !found {
		return
	}

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

	u, err := url.Parse(entryPoint)
	if err != nil {
		logger.WithE(err).Fatal("TRANSMISSION_API_ENTRYPOINT is not valid url", zap.String("value", entryPoint))
	}

	username, password := utils.GetUserPass(u.User)
	port := utils.GetPort(u)

	client, err := transmissionrpc.New(u.Hostname(), username, password, &transmissionrpc.AdvancedConfig{
		HTTPS: u.Scheme == "https",
		Port:  port,
	})
	if err != nil {
		logger.Fatal("failed to create transmission client")
	}

	router.Get("/transmission/metrics", createTransmissionHandler(client, interval))
}

func createTransmissionHandler(client *transmissionrpc.Client, interval time.Duration) fiber.Handler {
	var torrents []transmissionrpc.Torrent
	var status transmissionrpc.SessionStats
	var torrentMux, statusMux sync.RWMutex
	var torrentsErr, statusErr error

	var torrentFields = []string{"hashString", "status", "name", "labels", "uploadedEver", "downloadedEver"}
	var torrentFunc = func() {
		v, err := client.TorrentGet(context.TODO(), torrentFields, nil)
		torrentMux.Lock()
		defer torrentMux.Unlock()
		if err != nil {
			torrentsErr = errors.Wrap(err, "failed to get torrents")
			logger.WithE(torrentsErr).Error("failed to get torrents")
		} else {
			torrentsErr = nil
			torrents = v
		}
	}

	var statusFunc = func() {
		v, err := client.SessionStats(context.TODO())
		statusMux.Lock()
		defer statusMux.Unlock()
		if err != nil {
			statusErr = errors.Wrap(err, "failed to get session stats")
			logger.WithE(statusErr).Error("failed to get session stats")
		} else {
			statusErr = nil
			status = v
		}
	}

	logger.Info("start fetching transmission torrent details")
	go runInBackground(interval, torrentFunc)
	logger.Info("start fetching transmission global data")
	go runInBackground(interval, statusFunc)

	return func(ctx *fiber.Ctx) error {
		logger.Info("export transmission metrics")
		statusMux.RLock()
		if statusErr != nil {
			statusMux.RUnlock()

			return statusErr
		}

		fmt.Fprintln(ctx, "# without label filter")
		fmt.Fprintf(ctx, "transmission_download_all_total %d\n", status.CumulativeStats.DownloadedBytes)
		fmt.Fprintf(ctx, "transmission_upload_all_total %d\n", status.CurrentStats.UploadedBytes)
		statusMux.RUnlock()

		torrentMux.RLock()
		defer torrentMux.RUnlock()

		if torrentsErr != nil {
			return torrentsErr
		}

		fmt.Fprintln(ctx, "\n# all torrents")
		for i := range torrents {
			writeTorrent(ctx, &torrents[i])
		}

		return nil
	}
}

func runInBackground(interval time.Duration, f func()) {
	f()
	for range time.NewTicker(interval).C {
		f()
	}
}

func writeTorrent(w io.Writer, t *transmissionrpc.Torrent) {
	fmt.Fprintln(w, "\n# torrent", strconv.Quote(*t.Name))
	fmt.Fprintln(w, "# labels:", strings.Join(t.Labels, ", "))

	if len(t.Labels) == 0 {
		label := fmt.Sprintf("hash=%s, status=%s", strconv.Quote(*t.HashString), strconv.Quote(t.Status.String()))

		fmt.Fprintf(w, "transmission_torrent_download_bytes{%s} %d\n", label, *t.DownloadedEver)
		fmt.Fprintf(w, "transmission_torrent_upload_bytes{%s} %d\n", label, *t.UploadedEver)
	} else {
		for _, label := range t.Labels {
			fmt.Fprintln(w, "# label ", label)
			label := fmt.Sprintf("label=%s, hash=%s, status=%s",
				strconv.Quote(label), strconv.Quote(*t.HashString), strconv.Quote(t.Status.String()))

			fmt.Fprintf(w, "transmission_torrent_download_bytes{%s} %d\n", label, *t.DownloadedEver)
			fmt.Fprintf(w, "transmission_torrent_upload_bytes{%s} %d\n", label, *t.UploadedEver)
		}
	}
}
