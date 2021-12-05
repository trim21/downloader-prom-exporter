package handler

import (
	"fmt"
	"io"
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

	router.Get("/qbit/metrics", createQbitHandler(rpc))
}

func createQbitHandler(rpc *qbittorrent.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		success, err := rpc.Login("", "")
		if err != nil {
			return errors.Wrap(err, "failed to login")
		}

		if !success {
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

		writeGlobalData(ctx, &d.ServerState, t)

		// torrents, err := rpc.Torrents()
		// if err != nil {
		// 	return errors.Wrap(err, "failed to get torrents")
		// }

		for hash := range d.Torrents {
			writeQBitTorrent(ctx, hash, d.Torrents[hash])
		}

		return nil
	}
}

const qDefaultCategory = "UN-CATEGORIZED"

func writeGlobalData(w io.Writer, s *qbittorrent.ServerState, t *qbittorrent.Transfer) {
	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(s.AllTimeUl))
	fmt.Fprintf(w, "%s_upload_total_bytes %d\n\n", qPrefix, s.AllTimeUl)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(s.AllTimeDl))
	fmt.Fprintf(w, "%s_download_total_bytes %d\n\n", qPrefix, s.AllTimeDl)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(t.DlInfoData))
	fmt.Fprintf(w, "%s_dl_info_data_bytes %d\n\n", qPrefix, t.DlInfoData)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(t.UpInfoData))
	fmt.Fprintf(w, "%s_up_info_data_bytes %d\n\n", qPrefix, t.UpInfoData)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(int64(s.TotalBuffersSize)))
	fmt.Fprintf(w, "%s_total_buffers_size %d\n\n", qPrefix, s.TotalBuffersSize)

	fmt.Fprintf(w, "%s_dht_nodes %d\n", qPrefix, t.DhtNodes)
	fmt.Fprintf(w, "%s_read_cache_hits %s\n", qPrefix, s.ReadCacheHits)
	fmt.Fprintf(w, "%s_read_cache_overload %s\n", qPrefix, s.ReadCacheOverload)
	fmt.Fprintf(w, "%s_write_cache_overload %s\n", qPrefix, s.WriteCacheOverload)

	fmt.Fprintf(w, "%s_queued_io_jobs %d\n", qPrefix, s.QueuedIoJobs)
	fmt.Fprintf(w, "%s_average_queue_time_ms %d\n", qPrefix, s.AverageTimeQueue)
}

func writeQBitTorrent(w io.Writer, hash string, t qbittorrent.Torrent) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# torrent", strconv.Quote(t.Name))
	fmt.Fprintln(w, "# category:", t.Category)

	var label string
	if t.Category != "" {
		label = fmt.Sprintf("category=%s, hash=%s, state=%s",
			strconv.Quote(t.Category), strconv.Quote(hash), strconv.Quote(t.State))
	} else {
		label = fmt.Sprintf("category=%s, hash=%s, state=%s",
			strconv.Quote(qDefaultCategory), strconv.Quote(hash), strconv.Quote(t.State))
	}

	fmt.Fprintf(w, "%s_torrent_todo_bytes{%s} %d\n", qPrefix, label, t.AmountLeft)
	fmt.Fprintf(w, "%s_torrent_download_bytes{%s} %d\n", qPrefix, label, t.Downloaded)
	fmt.Fprintf(w, "%s_torrent_upload_bytes{%s} %d\n", qPrefix, label, t.Uploaded)
}
