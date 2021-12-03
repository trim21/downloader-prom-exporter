package rtorrent

import (
	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/pkg/errors"

	"app/pkg/utils"
)

const (
	DDownloadTotal rtorrent.Field = "d.down.total"
	DUploadTotal   rtorrent.Field = "d.up.total"
)

// Torrent represents a torrent in rTorrent
type Torrent struct {
	Name          string
	Hash          string
	Label         string
	DownloadTotal int
	UploadTotal   int
}

func (t Torrent) Labels() []string {
	return utils.SplitByComma(t.Label)
}

func GetTorrents(rpc *xmlrpc.Client) ([]Torrent, error) {
	args := []interface{}{
		"",
		string(rtorrent.ViewMain),
		rtorrent.DName.Query(),
		rtorrent.DHash.Query(),
		rtorrent.DLabel.Query(),
		DDownloadTotal.Query(),
		DUploadTotal.Query(),
	}
	results, err := rpc.Call("d.multicall2", args...)

	var torrents []Torrent
	if err != nil {
		return torrents, errors.Wrap(err, "d.multicall2 XMLRPC call failed")
	}

	for _, outerResult := range results.([]interface{}) {
		for _, innerResult := range outerResult.([]interface{}) {
			torrentData := innerResult.([]interface{})
			torrents = append(torrents, Torrent{
				Name:          torrentData[0].(string),
				Hash:          torrentData[1].(string),
				Label:         torrentData[2].(string),
				DownloadTotal: torrentData[3].(int),
				UploadTotal:   torrentData[4].(int),
			})
		}
	}

	return torrents, nil
}
