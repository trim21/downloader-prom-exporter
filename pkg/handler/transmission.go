package handler

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"

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

	var port uint16 = 9091

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

	fmt.Println(u.Hostname(), u.Host, port)

	const prefix = "transmission_"

	router.Get("/transmission/metrics", func(ctx *fiber.Ctx) error {
		client, err := transmissionrpc.New(u.Hostname(), username, password, &transmissionrpc.AdvancedConfig{
			HTTPS: u.Scheme == "https",
			Port:  port,
		})
		if err != nil {
			return err
		}

		torrents, err := client.TorrentGetAll(ctx.Context())
		if err != nil {
			return err
		}

		ctx.Status(200)

		labelDownloadCount := make(map[string]int64)
		labelUploadCount := make(map[string]int64)
		labelCount := make(map[string]int64)
		labelStatusCount := make(map[string]map[string]int64)

		status, err := client.SessionStats(ctx.Context())
		if err != nil {
			return errors.Wrap(err, "failed to get session stats")
		}

		for _, torrent := range torrents {
			for _, label := range torrent.Labels {
				labelDownloadCount[label] += *torrent.DownloadedEver
				labelUploadCount[label] += *torrent.UploadedEver
				labelCount[label]++
				labelStatusCount[label] = increase(labelStatusCount[label], torrent.Status.String())
			}
		}

		fmt.Fprintln(ctx, "# without label filter")
		fmt.Fprintf(ctx, "%sdownload_all_total %d\n", prefix, status.CumulativeStats.DownloadedBytes)
		fmt.Fprintf(ctx, "%supload_all_total %d\n", prefix, status.CurrentStats.UploadedBytes)

		fmt.Fprintln(ctx, "# download and upload label filter")
		fmt.Fprintln(ctx, "# some torrents are not included in this metrics")
		for _, label := range keys(labelDownloadCount) {
			fmt.Fprintf(ctx, "%sdownload_total{label=%s} %d\n", prefix, strconv.Quote(label), labelDownloadCount[label])
			fmt.Fprintf(ctx, "%supload_total{label=%s} %d\n", prefix, strconv.Quote(label), labelUploadCount[label])
			for _, status := range keys(labelStatusCount[label]) {
				fmt.Fprintf(ctx, "%scount{label=%s, status=%s} %d\n", prefix, strconv.Quote(label), strconv.Quote(status), labelCount[label])
			}
			fmt.Fprintln(ctx)
		}

		return nil
	})
}

func keys(m map[string]int64) []string {
	var s = make([]string, 0, len(m))

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
