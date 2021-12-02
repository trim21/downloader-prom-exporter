package handler

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mrobinsn/go-rtorrent/rtorrent"
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

	fmt.Println(entryPoint)

	conn := rtorrent.New(entryPoint, true)
	rpc := xmlrpc.NewClient(entryPoint, true)

	router.Get("/rtorrent/metrics", func(ctx *fiber.Ctx) error {
		v, err := getSummary(conn)
		if err != nil {
			return err
		}

		torrents, err := rtorrent2.GetTorrents(rpc, rtorrent.ViewSeeding)
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

func getSummary(conn *rtorrent.RTorrent) (*RTorrentTransSummary, error) {
	var err error
	v := &RTorrentTransSummary{}

	v.Hostname, err = conn.Name()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rTorrent down total")
	}

	v.DownTotal, err = conn.DownTotal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rTorrent down total")
	}

	v.UpTotal, err = conn.UpTotal()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rTorrent down total")
	}

	return v, nil
}
