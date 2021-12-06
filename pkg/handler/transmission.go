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
	"github.com/sirupsen/logrus"
)

func setupTransmissionMetrics(router fiber.Router) {
	entryPoint, found := os.LookupEnv("TRANSMISSION_API_ENTRYPOINT")
	if !found {
		return
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		logrus.Fatalf("TRANSMISSION_API_ENTRYPOINT %s is not valid url", entryPoint)
	}

	username := ""
	password := ""

	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	var port uint16

	r := u.Port()
	if r != "" {
		v, err := strconv.Atoi(r)
		if err != nil {
			logrus.Fatalln(v)
		}
		port = uint16(v)
	} else {
		if u.Scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}

	client, err := transmissionrpc.New(u.Hostname(), username, password, &transmissionrpc.AdvancedConfig{
		HTTPS: u.Scheme == "https",
		Port:  port,
	})
	if err != nil {
		logrus.Fatalln("failed to create transmission client")
	}

	router.Get(
		"/transmission/metrics",
		createTransmissionHandler(client),
	)
}

func createTransmissionHandler(client *transmissionrpc.Client) fiber.Handler {
	var torrents []transmissionrpc.Torrent
	var torrentMux sync.RWMutex
	var torrentsErr error

	var status transmissionrpc.SessionStats
	var statusMux sync.RWMutex
	var statusErr error

	var torrentFields = []string{"hashString", "status", "name", "labels", "uploadedEver", "downloadedEver"}

	var torrentFunc = func() {
		if v, err := client.TorrentGet(context.TODO(), torrentFields, nil); err != nil {
			torrentMux.Lock()
			torrentsErr = errors.Wrap(err, "failed to get torrents")
			logrus.Errorln(torrentsErr)
			torrentMux.Unlock()
		} else {
			torrentMux.Lock()
			torrentsErr = nil
			torrents = v
			torrentMux.Unlock()
		}
	}

	var statusFunc = func() {
		if v, err := client.SessionStats(context.TODO()); err != nil {
			statusMux.Lock()
			statusErr = errors.Wrap(err, "failed to get session stats")
			logrus.Errorln(statusErr)
			statusMux.Unlock()
		} else {
			statusMux.Lock()
			statusErr = nil
			status = v
			statusMux.Unlock()
		}
	}

	torrentFunc()
	statusFunc()

	go func() {
		for range time.NewTicker(time.Second * 5).C {
			torrentFunc()
		}
	}()

	go func() {
		for range time.NewTicker(time.Second * 5).C {
			statusFunc()
		}
	}()

	return func(ctx *fiber.Ctx) error {
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

		statusCount := make(map[string]int64)
		for _, torrent := range torrents {
			statusCount[torrent.Status.String()]++
		}

		for status, value := range statusCount {
			fmt.Fprintf(ctx, "transmission_download_all_count{status=%s} %d\n",
				strconv.Quote(status), value)
		}

		fmt.Fprintln(ctx, "\n# all torrents")
		for i := range torrents {
			writeTorrent(ctx, &torrents[i])
		}

		return nil
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
