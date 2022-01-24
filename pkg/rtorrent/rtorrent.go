package rtorrent

import (
	"encoding/xml"

	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/pkg/errors"

	"app/pkg/utils"
)

const (
	DDownloadTotal rtorrent.Field = "d.down.total"
	DUploadTotal   rtorrent.Field = "d.up.total"
)

// Torrent represents a torrent in rTorrent.
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

var ErrDecodeTorrent = xml.UnmarshalError("can't decode torrent property tuple")

func parseTorrents(results interface{}) ([]Torrent, error) {
	var torrents []Torrent

	for _, outerResult := range results.([]interface{}) {
		for _, innerResult := range outerResult.([]interface{}) {
			torrentData, ok := innerResult.([]interface{})
			if !ok {
				return nil, ErrDecodeTorrent
			}

			t := Torrent{}

			if t.Name, ok = torrentData[0].(string); !ok {
				return nil, ErrDecodeTorrent
			}

			if t.Hash, ok = torrentData[1].(string); !ok {
				return nil, ErrDecodeTorrent
			}

			if t.Label, ok = torrentData[2].(string); !ok {
				return nil, ErrDecodeTorrent
			}

			if t.DownloadTotal, ok = torrentData[3].(int); !ok {
				return nil, ErrDecodeTorrent
			}

			if t.UploadTotal, ok = torrentData[4].(int); !ok {
				return nil, ErrDecodeTorrent
			}

			torrents = append(torrents, t)
		}
	}

	return torrents, nil
}

type MainData struct {
	Hostname  string
	Torrents  []Torrent
	UpTotal   int
	DownTotal int
}

type call struct {
	MethodName string        `xml:"methodName"`
	Params     []interface{} `xml:"params"`
}

var ErrUnmarshal = xml.UnmarshalError("failed to decode xmlrpc multicall response")

func GetGlobalData(rpc *xmlrpc.Client) (*MainData, error) {
	results, err := rpc.Call("system.multicall", []call{
		{MethodName: "system.hostname", Params: nil},
		{MethodName: "throttle.global_down.total", Params: nil},
		{MethodName: "throttle.global_up.total", Params: nil},
		{MethodName: "d.multicall2", Params: []interface{}{
			"",
			string(rtorrent.ViewMain),
			rtorrent.DName.Query(),
			rtorrent.DHash.Query(),
			rtorrent.DLabel.Query(),
			DDownloadTotal.Query(),
			DUploadTotal.Query(),
		}},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current status")
	}

	v := &MainData{}

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
			[ [...], [...] ], //torrents
		]
		```
	*/

	if v.Hostname, ok = getString(r, 0); !ok {
		return nil, errors.Wrap(ErrUnmarshal, "failed to decode 'system.hostname'")
	}

	if v.DownTotal, ok = getInt(r, 1); !ok {
		return nil, errors.Wrap(ErrUnmarshal, "failed to decode 'throttle.global_down.total'")
	}

	if v.UpTotal, ok = getInt(r, 2); !ok { //nolint:gomnd
		return nil, errors.Wrap(ErrUnmarshal, "failed to decode 'throttle.global_up.total'")
	}

	if v.Torrents, err = parseTorrents(r[3]); err != nil {
		return nil, errors.Wrap(err, "failed to decode Torrents")
	}

	return v, nil
}

// get first value from [][]interface{} as string.
func getString(r []interface{}, index int) (string, bool) {
	vv, ok := r[index].([]interface{})
	if !ok {
		return "", ok
	}

	v, ok := vv[0].(string)

	return v, ok
}

// get first value from [][]interface{} as int.
func getInt(r []interface{}, index int) (int, bool) {
	vv, ok := r[index].([]interface{})
	if !ok {
		return 0, ok
	}

	v, ok := vv[0].(int)

	return v, ok
}
