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

		for hash := range d.Torrents {
			writeQBitTorrent(ctx, hash, d.Torrents[hash])
		}

		return nil
	}
}

const qDefaultCategory = "UN-CATEGORIZED"

func writeGlobalData(w io.Writer, s *qbittorrent.ServerState, t *qbittorrent.Transfer) {
	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(s.AllTimeUl))
	fmt.Fprintf(w, "qbittorrent_upload_total_bytes %d\n\n", s.AllTimeUl)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(s.AllTimeDl))
	fmt.Fprintf(w, "qbittorrent_download_total_bytes %d\n\n", s.AllTimeDl)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(t.DlInfoData))
	fmt.Fprintf(w, "qbittorrent_dl_info_data_bytes %d\n\n", t.DlInfoData)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(t.UpInfoData))
	fmt.Fprintf(w, "qbittorrent_up_info_data_bytes %d\n\n", t.UpInfoData)

	fmt.Fprintf(w, "# %s\n", utils.ByteCountIEC(int64(s.TotalBuffersSize)))
	fmt.Fprintf(w, "qbittorrent_total_buffers_size %d\n\n", s.TotalBuffersSize)

	fmt.Fprintf(w, "qbittorrent_dht_nodes %d\n", t.DhtNodes)
	fmt.Fprintf(w, "qbittorrent_read_cache_hits %s\n", s.ReadCacheHits)
	fmt.Fprintf(w, "qbittorrent_read_cache_overload %s\n", s.ReadCacheOverload)
	fmt.Fprintf(w, "qbittorrent_write_cache_overload %s\n", s.WriteCacheOverload)

	fmt.Fprintf(w, "qbittorrent_queued_io_jobs %d\n", s.QueuedIoJobs)
	fmt.Fprintf(w, "qbittorrent_average_queue_time_ms %d\n", s.AverageTimeQueue)
}

func writeQBitTorrent(w io.Writer, hash string, t qbittorrent.Torrent) {
	fmt.Fprintln(w, "\n# torrent", strconv.Quote(t.Name))
	fmt.Fprintln(w, "# category:", t.Category)

	var label string
	if t.Category != "" {
		label = fmt.Sprintf("category=%s, hash=%s, state=%s",
			strconv.Quote(t.Category), strconv.Quote(hash), strconv.Quote(t.State))
	} else {
		label = fmt.Sprintf("category=%s, hash=%s, state=%s",
			strconv.Quote(qDefaultCategory), strconv.Quote(hash), strconv.Quote(t.State))
	}

	var restUpload float64

	switch t.State {
	case qbittorrent.StateUploading, qbittorrent.StateStalledUploading, qbittorrent.StateDownloading:
		v := t.MaxRatio - t.Ratio
		if v > 0 {
			restUpload = float64(t.Downloaded) * v
		}
	case qbittorrent.StateCheckingUploading, qbittorrent.StateMoving:
		toUpload := float64(t.Size) * (t.MaxRatio - t.Ratio)
		uploaded := float64(t.Uploaded)
		if toUpload > uploaded {
			restUpload = toUpload - uploaded
		}
	}

	fmt.Fprintf(w, "qbittorrent_torrent_upload_todo_bytes{%s} %.1f\n", label, restUpload)
	switch t.State {
	case qbittorrent.StateCheckingUploading:
		fmt.Fprintf(w, "qbittorrent_torrent_todo_bytes{%s} 0\n", label)

	case qbittorrent.StatePausedUploading, qbittorrent.StatePausedDownloading, qbittorrent.StateForceDownloading:
		fmt.Fprintf(w, "qbittorrent_torrent_todo_bytes{%s} 0\n", label)

	default:
		fmt.Fprintf(w, "qbittorrent_torrent_todo_bytes{%s} %d\n", label, t.AmountLeft)
	}

	switch t.State {
	case qbittorrent.StateStalledUploading, qbittorrent.StateUploading,
		qbittorrent.StateForceUploading:
		fmt.Fprintf(w, "qbittorrent_torrent_seeding_bytes{%s} %d\n", label, t.Size)
	case qbittorrent.StateDownloading, qbittorrent.StateStalledDownloading,
		qbittorrent.StateForceDownloading:
		fmt.Fprintf(w, "qbittorrent_torrent_seeding_bytes{%s} %d\n", label, t.Downloaded)
	default:
		fmt.Fprintf(w, "qbittorrent_torrent_seeding_bytes{%s} %d\n", label, 0)
	}

	fmt.Fprintf(w, "qbittorrent_torrent_download_bytes{%s} %d\n", label, t.Downloaded)
	fmt.Fprintf(w, "qbittorrent_torrent_upload_bytes{%s} %d\n", label, t.Uploaded)
}
