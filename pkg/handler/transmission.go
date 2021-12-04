package handler

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const trPrefix = "transmission"

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
	return func(ctx *fiber.Ctx) error {
		client, err := transmissionrpc.New(hostname, username, password, &transmissionrpc.AdvancedConfig{
			HTTPS: scheme == "https",
			Port:  port,
		})
		if err != nil {
			return errors.Wrap(err, "failed to create transmission rpc client")
		}

		torrents, err := client.TorrentGetAll(ctx.Context())
		if err != nil {
			return errors.Wrap(err, "failed to get torrents")
		}

		status, err := client.SessionStats(ctx.Context())
		if err != nil {
			return errors.Wrap(err, "failed to get session stats")
		}

		labelStatusCount := make(map[string]map[string]int64)
		statusCount := make(map[string]int64)

		for _, torrent := range torrents {
			statusCount[torrent.Status.String()]++

			for _, label := range torrent.Labels {
				labelStatusCount[label] = increase(labelStatusCount[label], torrent.Status.String())
			}
		}

		fmt.Fprintln(ctx, "# without label filter")
		fmt.Fprintf(ctx, "%s_download_all_total %d\n", trPrefix, status.CumulativeStats.DownloadedBytes)
		fmt.Fprintf(ctx, "%s_upload_all_total %d\n", trPrefix, status.CurrentStats.UploadedBytes)

		for _, status := range keys(statusCount) {
			fmt.Fprintf(ctx, "%s_download_all_count{status=%s} %d\n",
				trPrefix, strconv.Quote(status), statusCount[status])
		}

		fmt.Fprintln(ctx, "# all torrents")

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
		label := fmt.Sprintf("hash=%s", strconv.Quote(*t.HashString))

		fmt.Fprintf(w, "%s_download_total{%s} %d\n", trPrefix, label, *t.DownloadedEver)

		fmt.Fprintf(w, "%s_upload_total{%s} %d\n", trPrefix, label, *t.UploadedEver)

		fmt.Fprintf(w, "%s_torrent_download_bytes{%s} %d\n", trPrefix, label, *t.DownloadedEver)

		fmt.Fprintf(w, "%s_torrent_upload_bytes{%s} %d\n", trPrefix, label, *t.UploadedEver)
	} else {
		for _, label := range t.Labels {
			label := fmt.Sprintf("label=%s, hash=%s", strconv.Quote(label), strconv.Quote(*t.HashString))

			fmt.Fprintf(w, "%s_download_total{%s} %d\n", trPrefix, label, *t.DownloadedEver)

			fmt.Fprintf(w, "%s_upload_total{%s} %d\n", trPrefix, label, *t.UploadedEver)

			fmt.Fprintf(w, "%s_torrent_download_bytes{%s} %d\n", trPrefix, label, *t.DownloadedEver)

			fmt.Fprintf(w, "%s_torrent_upload_bytes{%s} %d\n", trPrefix, label, *t.UploadedEver)
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

func increase(m map[string]int64, key string) map[string]int64 {
	if m == nil {
		m = make(map[string]int64)
	}
	m[key]++

	return m
}
