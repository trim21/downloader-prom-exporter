package handler

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Torrent struct {
	Link          string `xml:"link"`
	ContentLength int64  `xml:"contentLength"`
	PubDate       string `xml:"pubDate"`
}

var client = resty.New()
var tz = time.FixedZone("Asia/Shanghai", 8*60*60)

func SetupRouter(router fiber.Router) {
	setupRTorrentMetrics(router)
	setupQBitTorrentMetrics(router)
	setupTransmissionMetrics(router)

	router.Get("/rss/:protocol/:domain/+", resizeRss)

	router.Get("/mikanani/list.xml", func(c *fiber.Ctx) error {
		return rewriteRSS(c, "https://mikanani.me/RSS/Classic")
	})

	router.Get("/mikanani/bangumi", func(c *fiber.Ctx) error {
		var id = c.Query("id")
		if id == "" {
			return c.Status(fiber.StatusNotFound).SendString("")
		}

		q := url.Values{"bangumiId": []string{id}}
		if group := c.Query("subgroupid"); group != "" {
			q.Set("subgroupid", group)
		}

		return rewriteRSS(c, "https://mikanani.me/RSS/Bangumi?"+q.Encode())
	})
	router.Get("/debug", func(c *fiber.Ctx) error {
		m := make(map[string]string)
		c.Request().Header.VisitAll(func(key, value []byte) {
			m[utils.UnsafeString(key)] = utils.UnsafeString(value)
		})
		return c.Status(200).JSON(m)
	})
}

func rewriteRSS(c *fiber.Ctx, url string) error {
	// 重写mikanani的rss feed，在item上添加`pubDate`以支持sonarr
	res, err := client.R().Get(url)
	if err != nil {
		return err
	}
	data := Rss{}
	if err := xml.Unmarshal(res.Body(), &data); err != nil {
		return err
	}

	for _, item := range data.Channel.Items {
		pubDate, err := time.ParseInLocation("2006-01-02T15:04:05.999", item.Torrent.PubDate, tz)
		if err != nil {
			return err
		}
		item.Torrent.PubDate = pubDate.In(tz).String()
		item.PubDate = item.Torrent.PubDate
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationXMLCharsetUTF8)
	e := xml.NewEncoder(c)
	e.Indent("", "  ")
	return e.Encode(data)
}

func resizeRss(c *fiber.Ctx) error {
	logrus.Debug(c.Request().RequestURI())
	var protocol = c.Params("protocol")
	var domain = c.Params("domain")
	var path = c.Params("+")

	q := c.Query("limit", "50")
	var limit, err = strconv.Atoi(q)
	if err != nil {
		return errors.Wrapf(err, "limit %s is not valid int", q)
	}

	query := url.Values{}

	c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		k := utils.UnsafeString(key)
		if k != "limit" {
			query.Add(k, utils.UnsafeString(value))
		}
	})

	u := fmt.Sprintf("%s://%s/%s", protocol, domain, path)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	logrus.Infoln(u)
	res, err := client.R().Get(u)
	if err != nil {
		return err
	}

	data := Rss{}
	if err := xml.Unmarshal(res.Body(), &data); err != nil {
		return err
	}

	if len(data.Channel.Items) > limit {
		data.Channel.Items = data.Channel.Items[:limit-1]
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationXMLCharsetUTF8)
	e := xml.NewEncoder(c)
	e.Indent("", "  ")
	return e.Encode(data)
}
