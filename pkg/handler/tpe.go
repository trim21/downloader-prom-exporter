package handler

import "encoding/xml"

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Text        string `xml:",chardata"`
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Items       []*struct {
			Text string `xml:",chardata"`
			Guid struct {
				Text        string `xml:",chardata"`
				IsPermaLink string `xml:"isPermaLink,attr"`
			} `xml:"guid"`
			Link        string `xml:"link"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Torrent     struct {
				Text          string `xml:",chardata"`
				Xmlns         string `xml:"xmlns,attr"`
				Link          string `xml:"link"`
				ContentLength string `xml:"contentLength"`
				PubDate       string `xml:"pubDate"`
			} `xml:"torrent"`
			Enclosure struct {
				Text   string `xml:",chardata"`
				Type   string `xml:"type,attr"`
				Length string `xml:"length,attr"`
				URL    string `xml:"url,attr"`
			} `xml:"enclosure"`
		} `xml:"item"`
	} `xml:"channel"`
}
