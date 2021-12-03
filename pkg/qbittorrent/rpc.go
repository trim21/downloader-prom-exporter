package qbittorrent

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"app/pkg/utils"
)

var ErrConnectToDaemon = errors.New("Can't send http request to daemon")

// ErrBadResponse means that qbittorrent sent back an unexpected response
var ErrBadResponse = errors.New("received bad response")

type Client struct {
	h             *resty.Client
	Authenticated bool
}

func NewClient(entryPoint *url.URL) (*Client, error) {
	h := resty.New().SetBaseURL(entryPoint.String())

	return &Client{
		h: h,
	}, nil
}

func (c *Client) post(u string, data interface{}) (*resty.Response, error) {
	return c.h.R().SetBody(data).Post(u)
}

func (c *Client) get(u string) (*resty.Response, error) {
	return c.h.R().Get(u)
}

// Login logs you in to the qbittorrent client
// returns the current authentication status
func (c *Client) Login(username string, password string) (loggedIn bool, err error) {
	resp, err := c.h.R().
		SetQueryParam("username", username).
		SetQueryParam("password", password).
		Get("api/v2/auth/login")

	if err != nil {
		return false, err
	} else if resp.StatusCode() != 200 { // check for correct status code
		fmt.Println(resp.String())
		fmt.Println(resp.Request.URL)
		return false, errors.Wrap(ErrBadResponse, "couldn't log in")
	}

	// change authentication status so we know were authenticated in later requests
	c.Authenticated = true

	return c.Authenticated, nil
}

type Torrent struct {
	Name       string `json:"name"`
	Hash       string `json:"hash"`
	RawTags    string `json:"tags"`
	Uploaded   int64  `json:"uploaded"`
	Downloaded int64  `json:"downloaded"`
	Category   string `json:"category"`
}

func (t Torrent) Tags() []string {
	return utils.SplitByComma(t.RawTags)
}

func (c *Client) Torrents() ([]Torrent, error) {
	var t []Torrent
	resp, err := c.h.R().SetResult(&t).Get("api/v2/torrents/info")
	if err != nil {
		return nil, errors.Wrap(ErrConnectToDaemon, "")
	}

	if resp.StatusCode() >= 300 {
		logrus.Debugln(resp.String())
		return nil, errors.Wrap(ErrBadResponse, "status code >= 300")
	}

	return t, nil
}

type MainData struct {
	FullUpdate  bool `json:"full_update"`
	Rid         int  `json:"rid"`
	ServerState struct {
		AllTimeDl            int64  `json:"alltime_dl"`
		AllTimeUl            int64  `json:"alltime_ul"`
		AverageTimeQueue     int    `json:"average_time_queue"`
		ConnectionStatus     string `json:"connection_status"`
		DhtNodes             int    `json:"dht_nodes"`
		DlInfoData           int64  `json:"dl_info_data"`
		DlInfoSpeed          int    `json:"dl_info_speed"`
		DlRateLimit          int    `json:"dl_rate_limit"`
		FreeSpaceOnDisk      int64  `json:"free_space_on_disk"`
		GlobalRatio          string `json:"global_ratio"`
		QueuedIoJobs         int    `json:"queued_io_jobs"`
		Queueing             bool   `json:"queueing"`
		ReadCacheHits        string `json:"read_cache_hits"`
		ReadCacheOverload    string `json:"read_cache_overload"`
		RefreshInterval      int    `json:"refresh_interval"`
		TotalBuffersSize     int    `json:"total_buffers_size"`
		TotalPeerConnections int    `json:"total_peer_connections"`
		TotalQueuedSize      int    `json:"total_queued_size"`
		TotalWastedSession   int    `json:"total_wasted_session"`
		UpInfoData           int64  `json:"up_info_data"`
		UpInfoSpeed          int    `json:"up_info_speed"`
		UpRateLimit          int    `json:"up_rate_limit"`
		UseAltSpeedLimits    bool   `json:"use_alt_speed_limits"`
		WriteCacheOverload   string `json:"write_cache_overload"`
	} `json:"server_state"`
}

func (c *Client) MainData() (*MainData, error) {
	var t = &MainData{}
	resp, err := c.h.R().SetResult(t).
		SetQueryParam("rid", "0").
		Get("api/v2/sync/maindata")

	if err != nil {
		return nil, errors.Wrap(err, "can't connect to http daemon")
	}

	if resp.StatusCode() >= 300 {
		logrus.Debugln(resp.String())
		return nil, errors.Wrap(ErrBadResponse, "status code >= 300")
	}

	return t, nil
}
