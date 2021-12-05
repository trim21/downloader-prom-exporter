package handler

import (
	"encoding/xml"
	"fmt"
	"io"
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
		v, err := getGlobalData(rpc)
		if err != nil {
			return err
		}

		torrents, err := rtorrent2.GetTorrents(rpc)
		if err != nil {
			return errors.Wrap(err, "failed to get torrents from rpc")
		}

		fmt.Fprintf(ctx, "rtorrent_upload_total_bytes{hostname=%s} %d\n", strconv.Quote(v.Hostname), v.UpTotal)
		fmt.Fprintf(ctx, "rtorrent_download_total_bytes{hostname=%s} %d\n", strconv.Quote(v.Hostname), v.DownTotal)

		for i := range torrents {
			writeRtorrentTorrent(ctx, &torrents[i])
		}

		return nil
	})
}

func writeRtorrentTorrent(w io.Writer, t *rtorrent2.Torrent) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# torrent", strconv.Quote(t.Name))
	fmt.Fprintln(w, "# label:", t.Label)

	if t.Label == "" {
		fmt.Fprintf(w, "rtorrent_torrent_download_total_bytes{hash=%s} %d\n",
			strconv.Quote(t.Hash), t.DownloadTotal)
	} else {
		for _, label := range t.Labels() {
			fmt.Fprintf(w, "rtorrent_torrent_download_bytes{label=%s, hash=%s} %d\n",
				strconv.Quote(label), strconv.Quote(t.Hash), t.DownloadTotal)

			fmt.Fprintf(w, "rtorrent_torrent_upload_bytes{label=%s, hash=%s} %d\n",
				strconv.Quote(label), strconv.Quote(t.Hash), t.UploadTotal)
		}
	}
}

type RTorrentTransSummary struct {
	Hostname  string
	UpTotal   int
	DownTotal int
}

var ErrUnmarshal = xml.UnmarshalError("failed to decode xmlrpc multicall response")

func getGlobalData(rpc *xmlrpc.Client) (*RTorrentTransSummary, error) {
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

	r1, ok := results.([]interface{})
	if !ok {
		return nil, ErrUnmarshal
	}

	r, ok := r1[0].([]interface{})
	if !ok {
		return nil, ErrUnmarshal
	}

	/*
		r is a value like this
		```
		[
			[	"hostname" ], // "system.hostname"
			[ 1 ], // "throttle.global_down.total"
			[ 2 ], // "throttle.global_up.total"
		]
		```
	*/

	v.Hostname, ok = getString(r, 0)
	if !ok {
		return nil, errors.Wrap(ErrUnmarshal, "failed to decode 'system.hostname'")
	}

	v.DownTotal, ok = getInt(r, 1)
	if !ok {
		return nil, errors.Wrap(ErrUnmarshal, "failed to decode 'throttle.global_down.total'")
	}

	v.UpTotal, ok = getInt(r, 2) //nolint:gomnd
	if !ok {
		return nil, errors.Wrap(ErrUnmarshal, "failed to decode 'throttle.global_up.total'")
	}

	return v, nil
}

// get first value from [][]interface{} as string
func getString(r []interface{}, index int) (string, bool) {
	vv, ok := r[index].([]interface{})
	if !ok {
		return "", ok
	}

	v, ok := vv[0].(string)

	return v, ok
}

// get first value from [][]interface{} as int
func getInt(r []interface{}, index int) (int, bool) {
	vv, ok := r[index].([]interface{})
	if !ok {
		return 0, ok
	}

	v, ok := vv[0].(int)

	return v, ok
}

// get first value from [][]interface{} as int
func getInt64(r []interface{}, index int) (int64, bool) {
	vv, ok := r[index].([]interface{})
	if !ok {
		return 0, ok
	}

	v, ok := vv[0].(int64)

	if !ok {
		fmt.Println(vv[0])
	}

	return v, ok
}
