package handler

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hekmon/transmissionrpc/v2"
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

	router.Get(
		"/transmission/metrics",
		createTransmissionHandler(u.Scheme, u.Hostname(), port, username, password),
	)
}

func createTransmissionHandler(scheme, hostname string, port uint16, username, password string) fiber.Handler {
	client, err := transmissionrpc.New(hostname, username, password, &transmissionrpc.AdvancedConfig{
		HTTPS: scheme == "https",
		Port:  port,
	})
	if err != nil {
		logrus.Fatalln("failed to create transmission client")
	}

	var torrents []transmissionrpc.Torrent
	var torrentMux sync.RWMutex

	var status transmissionrpc.SessionStats
	var statusMux sync.RWMutex

	var torrentFunc = func() {
		if v, err := client.TorrentGetAll(context.TODO()); err != nil {
			logrus.Errorln("failed to get torrents", err)
		} else {
			torrentMux.Lock()
			torrents = v
			torrentMux.Unlock()
		}
	}

	var statusFunc = func() {
		if v, err := client.SessionStats(context.TODO()); err != nil {
			logrus.Errorln("failed to get session stats", err)
		} else {
			statusMux.Lock()
			status = v
			statusMux.Unlock()
		}
	}

	torrentFunc()
	statusFunc()

	go func() {
		for range time.NewTimer(time.Second * 5).C {
			torrentFunc()
		}
	}()

	go func() {
		for range time.NewTimer(time.Second * 5).C {
			statusFunc()
		}
	}()

	return func(ctx *fiber.Ctx) error {
		statusMux.RLock()
		fmt.Fprintln(ctx, "# without label filter")
		fmt.Fprintf(ctx, "transmission_download_all_total %d\n", status.CumulativeStats.DownloadedBytes)
		fmt.Fprintf(ctx, "transmission_upload_all_total %d\n", status.CurrentStats.UploadedBytes)
		statusMux.RUnlock()

		torrentMux.RLock()

		statusCount := make(map[string]int64)
		for _, torrent := range torrents {
			statusCount[torrent.Status.String()]++
		}

		for _, status := range keys(statusCount) {
			fmt.Fprintf(ctx, "transmission_download_all_count{status=%s} %d\n",
				strconv.Quote(status), statusCount[status])
		}

		fmt.Fprintln(ctx, "\n# all torrents")

		for i := range torrents {
			writeTorrent(ctx, &torrents[i])
		}

		torrentMux.RUnlock()

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

func keys(m map[string]int64) []string {
	s := make([]string, 0, len(m))

	for label := range m {
		s = append(s, label)
	}

	sort.Strings(s)

	return s
}
