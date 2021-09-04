package handler

import (
	"encoding/xml"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
)

type Torrent struct {
	Link          string `xml:"link"`
	ContentLength int64  `xml:"contentLength"`
	PubDate       string `xml:"pubDate"`
}

var client = resty.New()

func SetupRouter(router fiber.Router) {

	router.Get("/mikanani/list.xml", func(c *fiber.Ctx) error {
		// 重写mikanani的rss feed，在item上添加`pubDate`以支持sonarr
		res, err := client.R().Get("https://mikanani.me/RSS/Classic")
		if err != nil {
			return err
		}
		data := Rss{}
		if err := xml.Unmarshal(res.Body(), &data); err != nil {
			return err
		}

		for _, item := range data.Channel.Items {
			pubDate, err := time.ParseInLocation("2006-01-02T15:04:05.999", item.Torrent.PubDate, time.Local)
			if err != nil {
				return err
			}
			item.Torrent.PubDate = pubDate.Local().String()
			item.PubDate = item.Torrent.PubDate
		}

		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationXMLCharsetUTF8)
		e := xml.NewEncoder(c)
		e.Indent("", "  ")
		return e.Encode(data)
	})
}
