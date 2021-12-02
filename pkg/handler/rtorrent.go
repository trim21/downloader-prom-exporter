package handler

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	rtorrent2 "app/pkg/rtorrent"
)

func setupRTorrentMetrics(router fiber.Router) {
	os.Setenv("RTORRENT_API_ENTRYPOINT", "https://rtorrent.omv.trim21.me/RPC2")

	entryPoint, found := os.LookupEnv("RTORRENT_API_ENTRYPOINT")
	if !found {
		return
	}

	_, err := url.Parse(entryPoint)
	if err != nil {
		logrus.Fatalf("can't parse RTORRENT_API_ENTRYPOINT %s", entryPoint)
	}

	rpc := xmlrpc.NewClient(entryPoint, true)

	router.Get("/rtorrent/metrics", func(ctx *fiber.Ctx) error {
		v, err := getSummary2(rpc)
		if err != nil {
			return err
		}

		torrents, err := rtorrent2.GetTorrents(rpc)
		if err != nil {
			return errors.Wrap(err, "failed to get torrents from rpc")
		}

		fmt.Fprintf(ctx, "rtorrent_upload_total_bytes{hostname=%s} %d\n", strconv.Quote(v.Hostname), v.UpTotal)
		fmt.Fprintf(ctx, "rtorrent_download_total_bytes{hostname=%s} %d\n", strconv.Quote(v.Hostname), v.DownTotal)

		for _, torrent := range torrents {
			writeRtorrentTorrent(ctx, &torrent)
		}

		return nil
	})
}

const rPrefix = "rtorrent"

func writeRtorrentTorrent(ctx *fiber.Ctx, t *rtorrent2.Torrent) {
	fmt.Fprintln(ctx)
	fmt.Fprintln(ctx, "# torrent", strconv.Quote(t.Name))
	fmt.Fprintln(ctx, "# label:", t.Label)

	if t.Label == "" {
		fmt.Fprintf(ctx,
			"%s_torrent_download_total_bytes{hash=%s} %d\n",
			rPrefix, strconv.Quote(t.Hash), t.DownloadTotal)
	} else {
		for _, label := range t.Labels() {
			fmt.Fprintf(ctx,
				"%s_torrent_download_bytes{label=%s, hash=%s} %d\n",
				rPrefix, strconv.Quote(label), strconv.Quote(t.Hash), t.DownloadTotal)

			fmt.Fprintf(ctx,
				"%s_torrent_upload_bytes{label=%s, hash=%s} %d\n",
				rPrefix, strconv.Quote(label), strconv.Quote(t.Hash), t.UploadTotal)
		}
	}
}

type RTorrentTransSummary struct {
	Hostname  string
	UpTotal   int
	DownTotal int
}

func getSummary2(rpc *xmlrpc.Client) (*RTorrentTransSummary, error) {
	results, err := rpc.Call("system.multicall", []interface{}{
		map[string]interface{}{
			"methodName": "system.hostname",
			"params":     []string{},
		},
		map[string]interface{}{
			"methodName": "throttle.global_down.total",
			"params":     []string{},
		},
		map[string]interface{}{
			"methodName": "throttle.global_up.total",
			"params":     []string{},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get current status")
	}

	v := &RTorrentTransSummary{}

	r1 := (results.([]interface{}))[0]
	r2 := r1.([]interface{})

	r := r2

	v.Hostname = r[0].([]interface{})[0].(string)
	v.DownTotal = r[1].([]interface{})[0].(int)
	v.UpTotal = r[2].([]interface{})[0].(int)

	return v, nil
}
