package handler

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"app/pkg/qbittorrent"
	"app/pkg/utils"
)

const qPrefix = "qbittorrent"

func setupQBitMetrics(router fiber.Router) {
	entryPoint, found := os.LookupEnv("QBIT_API_ENTRYPOINT")
	if !found {
		return
	}

	u, err := url.Parse(entryPoint)
	if err != nil {
		logrus.Fatalf("can't parse QBIT_API_ENTRYPOINT %s", entryPoint)
	}

	rpc, err := qbittorrent.NewClient(u)
	if err != nil {
		logrus.Fatalln(err)
	}

	router.Get("/qbit/metrics", func(ctx *fiber.Ctx) error {
		if rpc == nil {
			return ctx.SendString("hehe")
		}
		logined, err := rpc.Login("", "")
		if err != nil {
			return errors.Wrap(err, "failed to login")
		}
		if !logined {
			return fiber.ErrUnauthorized
		}

		t, err := rpc.Transfer()
		if err != nil {
			return errors.Wrap(err, "failed to get transfer info")
		}

		d, err := rpc.MainData()
		if err != nil {
			return errors.Wrap(err, "failed to get main data")
		}

		fmt.Fprintf(ctx, "# %s\n", utils.ByteCountIEC(d.ServerState.AllTimeUl))
		fmt.Fprintf(ctx, "%s_upload_total_bytes %d\n\n", qPrefix, d.ServerState.AllTimeUl)

		fmt.Fprintf(ctx, "# %s\n", utils.ByteCountIEC(d.ServerState.AllTimeDl))
		fmt.Fprintf(ctx, "%s_download_total_bytes %d\n\n", qPrefix, d.ServerState.AllTimeDl)

		fmt.Fprintf(ctx, "# %s\n", utils.ByteCountIEC(t.DlInfoData))
		fmt.Fprintf(ctx, "%s_dl_info_data_bytes %d\n\n", qPrefix, t.DlInfoData)

		fmt.Fprintf(ctx, "# %s\n", utils.ByteCountIEC(t.UpInfoData))
		fmt.Fprintf(ctx, "%s_up_info_data_bytes %d\n\n", qPrefix, t.UpInfoData)

		fmt.Fprintf(ctx, "# %s\n", utils.ByteCountIEC(int64(d.ServerState.TotalBuffersSize)))
		fmt.Fprintf(ctx, "%s_total_buffers_size %d\n\n", qPrefix, d.ServerState.TotalBuffersSize)

		fmt.Fprintf(ctx, "%s_dht_nodes %d\n", qPrefix, t.DhtNodes)
		fmt.Fprintf(ctx, "%s_read_cache_hits %s\n", qPrefix, d.ServerState.ReadCacheHits)
		fmt.Fprintf(ctx, "%s_read_cache_overload %s\n", qPrefix, d.ServerState.ReadCacheOverload)
		fmt.Fprintf(ctx, "%s_write_cache_overload %s\n", qPrefix, d.ServerState.WriteCacheOverload)

		torrents, err := rpc.Torrents()
		if err != nil {
			return errors.Wrap(err, "failed to get torrents")
		}

		for _, t := range torrents {
			writeQBitTorrent(ctx, &t)
			// fmt.Fprintln(ctx)
			// fmt.Fprintln(ctx, t.Name)
			// fmt.Fprintln(ctx, t.Category)
		}

		return nil
	})
}

func writeQBitTorrent(ctx *fiber.Ctx, t *qbittorrent.Torrent) {
	fmt.Fprintln(ctx)
	fmt.Fprintln(ctx, "# torrent", strconv.Quote(t.Name))
	fmt.Fprintln(ctx, "# category:", t.Category)
	if t.Category != "" {

		fmt.Fprintf(ctx,
			"%s_torrent_download_bytes{category=%s, hash=%s} %d\n",
			qPrefix, strconv.Quote(t.Category), strconv.Quote(t.Hash), t.Downloaded)

		fmt.Fprintf(ctx,
			"%s_torrent_upload_bytes{category=%s, hash=%s} %d\n",
			qPrefix, strconv.Quote(t.Category), strconv.Quote(t.Hash), t.Uploaded)

	} else {
		fmt.Fprintf(ctx,
			"%s_torrent_download_bytes{hash=%s} %d\n",
			qPrefix, strconv.Quote(t.Hash), t.Downloaded)

		fmt.Fprintf(ctx,
			"%s_torrent_upload_bytes{hash=%s} %d\n",
			qPrefix, strconv.Quote(t.Hash), t.Uploaded)
	}
}
