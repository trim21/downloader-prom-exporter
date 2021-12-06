package qbittorrent

import (
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"app/pkg/utils"
)

const (
	StateChecking = "checkingUP"
	StateMoving   = "moving"

	StateUploading        = "uploading"
	StateStalledUploading = "stalledUP"

	StateDownloading        = "downloading"
	StateStalledDownloading = "stalledDL"

	StatePausedUploading   = "pausedUP"
	StatePausedDownloading = "pausedDL"
)

var ErrConnectToDaemon = errors.New("Can't finish http request")

// ErrBadResponse means that qbittorrent sent back an unexpected response.
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

// Login logs you in to the qbittorrent client
// returns the current authentication status.
func (c *Client) Login(username string, password string) (loggedIn bool, err error) {
	resp, err := c.h.R().
		SetQueryParam("username", username).
		SetQueryParam("password", password).
		Get("api/v2/auth/login")

	if err != nil {
		return false, errors.Wrap(ErrConnectToDaemon, err.Error())
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
	RawTags    string `json:"tags"`
	Category   string `json:"category"`
	State      string `json:"state"`
	Uploaded   int64  `json:"uploaded"`
	Downloaded int64  `json:"downloaded"`
	// download bytes
	Completed int64 `json:"completed"`
	// rest need to download bytes
	AmountLeft int64 `json:"amount_left"`

	// selected size
	Size int64 `json:"size"`
	// torrent content total size
	TotalSize    int64   `json:"total_size"`
	Progress     float64 `json:"progress"`
	SuperSeeding bool    `json:"super_seeding"`
	Ratio        float64 `json:"ratio"`
	MaxRatio     float64 `json:"max_ratio"`
}

// Hash       string `json:"hash"`

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

type ServerState struct {
	ReadCacheHits      string `json:"read_cache_hits"`
	WriteCacheOverload string `json:"write_cache_overload"`
	ReadCacheOverload  string `json:"read_cache_overload"`
	DhtNodes           int    `json:"dht_nodes"`
	DlInfoSpeed        int    `json:"dl_info_speed"`
	DlRateLimit        int    `json:"dl_rate_limit"`
	AverageTimeQueue   int    `json:"average_time_queue"`
	AllTimeUl          int64  `json:"alltime_ul"` //nolint:misspell
	DlInfoData         int64  `json:"dl_info_data"`
	UpInfoData         int64  `json:"up_info_data"`
	QueuedIoJobs       int    `json:"queued_io_jobs"`
	TotalBuffersSize   int    `json:"total_buffers_size"`
	AllTimeDl          int64  `json:"alltime_dl"` //nolint:misspell
}

// ConnectionStatus     string `json:"connection_status"`
// FreeSpaceOnDisk      int64  `json:"free_space_on_disk"`
// GlobalRatio          string `json:"global_ratio"`
// Queueing             bool   `json:"queueing"`
// RefreshInterval      int    `json:"refresh_interval"`
// TotalPeerConnections int    `json:"total_peer_connections"`
// TotalQueuedSize      int    `json:"total_queued_size"`
// TotalWastedSession   int    `json:"total_wasted_session"`
// UpInfoSpeed          int    `json:"up_info_speed"`
// UpRateLimit          int    `json:"up_rate_limit"`
// UseAltSpeedLimits    bool   `json:"use_alt_speed_limits"`

type MainData struct {
	Torrents    map[string]Torrent `json:"torrents"`
	ServerState ServerState        `json:"server_state"`
	FullUpdate  bool               `json:"full_update"`
	Rid         int                `json:"rid"`
}

type Transfer struct {
	DhtNodes   int   `json:"dht_nodes"`
	DlInfoData int64 `json:"dl_info_data"`
	UpInfoData int64 `json:"up_info_data"`
}

func (c *Client) Transfer() (*Transfer, error) {
	t := &Transfer{}

	resp, err := c.h.R().SetResult(t).Get("api/v2/transfer/info")
	if err != nil {
		return nil, errors.Wrap(err, "can't connect to http daemon")
	}

	if resp.StatusCode() >= 300 {
		return nil, errors.Wrap(ErrBadResponse, "status code >= 300")
	}

	return t, nil
}

func (c *Client) MainData() (*MainData, error) {
	t := &MainData{}

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
