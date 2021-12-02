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

const trPrefix = "transmission_"

func setupTransmissionMetrics(router fiber.Router) {

	os.Setenv("TRANSMISSION_API_ENTRYPOINT", "https://admin:password@bt.omv.trim21.me")

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

		labelStatusCount := make(map[string]map[string]int64)
		statusCount := make(map[string]int64)

		status, err := client.SessionStats(ctx.Context())
		if err != nil {
			return errors.Wrap(err, "failed to get session stats")
		}

		for _, torrent := range torrents {
			statusCount[torrent.Status.String()]++

			for _, label := range torrent.Labels {
				labelStatusCount[label] = increase(labelStatusCount[label], torrent.Status.String())
			}
		}

		fmt.Fprintln(ctx, "# without label filter")
		fmt.Fprintf(ctx, "%sdownload_all_total %d\n", trPrefix, status.CumulativeStats.DownloadedBytes)
		fmt.Fprintf(ctx, "%supload_all_total %d\n", trPrefix, status.CurrentStats.UploadedBytes)

		for _, status := range keys(statusCount) {
			fmt.Fprintf(ctx, "%sdownload_all_count{status=%s} %d\n", trPrefix, strconv.Quote(status), statusCount[status])
		}

		fmt.Fprintln(ctx, "# all torrents")
		for _, torrent := range torrents {
			writeTorrent(ctx, &torrent)
		}
		return nil
	})
}

func writeTorrent(ctx *fiber.Ctx, t *transmissionrpc.Torrent) {
	fmt.Fprintln(ctx, "# torrent", strconv.Quote(*t.Name))
	fmt.Fprintln(ctx, "# labels", t.Labels)

	if len(t.Labels) == 0 {
		fmt.Fprintf(ctx,
			"%sdownload_total{hash=%s} %d\n",
			trPrefix, strconv.Quote(*t.HashString), *t.DownloadedEver)

		fmt.Fprintf(ctx,
			"%supload_total{hash=%s} %d\n",
			trPrefix, strconv.Quote(*t.HashString), *t.UploadedEver)

	} else {
		for _, label := range t.Labels {
			fmt.Fprintf(ctx,
				"%sdownload_total{label=%s,hash=%s} %d\n",
				trPrefix, strconv.Quote(label), strconv.Quote(*t.HashString), *t.DownloadedEver)

			fmt.Fprintf(ctx,
				"%supload_total{label=%s,hash=%s} %d\n",
				trPrefix, strconv.Quote(label), strconv.Quote(*t.HashString), *t.UploadedEver)
		}
	}
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
