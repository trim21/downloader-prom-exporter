package rtorrent

import (
	"encoding/xml"

	"github.com/mrobinsn/go-rtorrent/rtorrent"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/trim21/errgo"

	"app/pkg/utils"
)

const (
	DDownloadTotal rtorrent.Field = "d.down.total"
	DUploadTotal   rtorrent.Field = "d.up.total"
)

// Torrent represents a torrent in rTorrent.
type Torrent struct {
	Name           string
	Hash           string
	Label          string
	DownloadTotal  int
	UploadTotal    int
	PeerConnecting int
}

func (t Torrent) Labels() []string {
	return utils.SplitByComma(t.Label)
}

var ErrDecodeTorrent = xml.UnmarshalError("can't decode torrent property tuple")

func parseTorrents(results any) ([]Torrent, error) {
	var torrents []Torrent

	for _, outerResult := range results.([]any) {
		for _, innerResult := range outerResult.([]any) {
			torrentData, ok := innerResult.([]any)
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

			if t.PeerConnecting, ok = torrentData[5].(int); !ok {
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
	DHTNodes  int
}

type call struct {
	MethodName string `xml:"methodName"`
	Params     []any  `xml:"params"`
}

var ErrUnmarshal = xml.UnmarshalError("failed to decode xmlrpc multicall response")

func GetGlobalData(rpc *xmlrpc.Client) (*MainData, error) {
	results, err := rpc.Call("system.multicall", []call{
		{MethodName: "system.hostname"},
		{MethodName: "throttle.global_down.total"},
		{MethodName: "throttle.global_up.total"},
		{MethodName: "d.multicall2", Params: []any{
			"",
			string(rtorrent.ViewMain),
			rtorrent.DName.Query(),
			rtorrent.DHash.Query(),
			rtorrent.DLabel.Query(),
			DDownloadTotal.Query(),
			DUploadTotal.Query(),
			"d.peers_connected=",
		}},
		{MethodName: "dht.statistics"},
	})
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get current status")
	}

	v := &MainData{}

	r1, ok := results.([]any)
	if !ok {
		return nil, ErrUnmarshal
	}

	r, ok := r1[0].([]any)
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
		return nil, errgo.Wrap(ErrUnmarshal, "failed to decode 'system.hostname'")
	}

	if v.DownTotal, ok = getInt(r, 1); !ok {
		return nil, errgo.Wrap(ErrUnmarshal, "failed to decode 'throttle.global_down.total'")
	}

	if v.UpTotal, ok = getInt(r, 2); !ok { //nolint:gomnd
		return nil, errgo.Wrap(ErrUnmarshal, "failed to decode 'throttle.global_up.total'")
	}

	if v.Torrents, err = parseTorrents(r[3]); err != nil {
		return nil, errgo.Wrap(err, "failed to decode Torrents")
	}

	v.DHTNodes = r[4].([]any)[0].(map[string]any)["active"].(int)

	return v, nil
}

// get first value from [][]any as string.
func getString(r []any, index int) (string, bool) {
	vv, ok := r[index].([]any)
	if !ok {
		return "", ok
	}

	v, ok := vv[0].(string)

	return v, ok
}

// get first value from [][]any as int.
func getInt(r []any, index int) (int, bool) {
	vv, ok := r[index].([]any)
	if !ok {
		return 0, ok
	}

	v, ok := vv[0].(int)

	return v, ok
}
