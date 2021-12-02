package handler

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
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

	fmt.Println(entryPoint)

	raw := `
flood_upload_total{hostname=[Name]} [UpTotal]
flood_download_total{hostname=[Name]} [DownTotal]
flood_upload_rate{hostname=[Name]} [UpRate]
flood_download_rate{hostname=[Name]} [DownRate]

`
	template := fasttemplate.New(strings.TrimSpace(raw), "[", "]")

	conn := rtorrent.New(entryPoint, true)
	router.Get("/rtorrent/metrics", func(ctx *fiber.Ctx) error {
		v, err := getSummary(conn)
		if err != nil {
			return err
		}

		s := template.ExecuteString(map[string]interface{}{
			"Name":      strconv.Quote(v.Hostname),
			"UpRate":    strconv.Itoa(v.UpRate),
			"UpTotal":   strconv.Itoa(v.UpTotal),
			"DownRate":  strconv.Itoa(v.DownRate),
			"DownTotal": strconv.Itoa(v.DownTotal),
		})
		return ctx.SendString(s)
	})
}

type RTorrentTransSummary struct {
	Hostname  string
	UpRate    int
	UpTotal   int
	DownRate  int
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

	v.UpRate, err = conn.UpRate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rTorrent down total")
	}

	v.DownRate, err = conn.DownRate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rTorrent down total")
	}

	return v, nil

}
