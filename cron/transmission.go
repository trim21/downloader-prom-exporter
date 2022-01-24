package cron

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-resty/resty/v2"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/robfig/cron/v3"
	"github.com/scylladb/go-set"
	"github.com/scylladb/go-set/strset"
	"go.uber.org/zap"

	"app/pkg/errgo"
	"app/pkg/logger"
)

var labelConfig = make(map[string]string) //nolint:gochecknoglobals

//nolint:gochecknoinits
func init() {
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "LABEL_") {
			pair := strings.SplitN(e, "=", 2)
			labelConfig[strings.ToLower(strings.TrimPrefix(pair[0], "LABEL_"))] = pair[1]
		}
	}
	logger.Info("label config", zap.Any("value", labelConfig))
}

func processLabels(rpc *transmissionrpc.Client, torrent transmissionrpc.Torrent) error {
	var labelExpected = strset.New()
	var currentLabels = strset.New(torrent.Labels...)
	var managedLabels = strset.NewWithSize(len(labelConfig))
	for label := range labelConfig {
		managedLabels.Add(label)
	}

	for label, dir := range labelConfig {
		if strings.HasPrefix(*torrent.DownloadDir+"/", dir) {
			labelExpected.Add(label)
		}
	}

	var expected = strset.Union(strset.Difference(currentLabels, managedLabels), labelExpected)
	if expected.IsEqual(currentLabels) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	payload := transmissionrpc.TorrentSetPayload{IDs: []int64{*torrent.ID}, Labels: expected.List()}
	if err := rpc.TorrentSet(ctx, payload); err != nil {
		logger.Error("rpc payload", zap.Stringp("name", torrent.Name), zap.Any("payload", payload))

		return errgo.Wrap(err, "rpc")
	}

	return nil
}

func processTracker(
	rpc *transmissionrpc.Client,
	trackers *strset.Set,
	m *sync.RWMutex,
	torrent transmissionrpc.Torrent,
) error {
	currentTrackers := set.NewStringSetWithSize(len(torrent.Trackers))
	trackersToAdd := set.NewStringSet()
	for _, tracker := range torrent.Trackers {
		currentTrackers.Add(tracker.Announce)
	}

	m.RLock()
	trackers.Each(func(tracker string) bool {
		if !currentTrackers.Has(tracker) {
			trackersToAdd.Add(tracker)
		}

		return true
	})
	m.RUnlock()

	if trackersToAdd.IsEmpty() {
		// nothing to do, skip
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	payload := transmissionrpc.TorrentSetPayload{IDs: []int64{*torrent.ID}, TrackerAdd: trackersToAdd.List()}
	if err := rpc.TorrentSet(ctx, payload); err != nil {
		logger.Error("rpc payload", zap.Stringp("name", torrent.Name), zap.Reflect("payload", payload))

		return errgo.Wrap(err, "rpc")
	}

	return nil
}

func setupTransmissionMetrics(rpc *transmissionrpc.Client, c *cron.Cron) error {
	if rpc == nil {
		return nil
	}

	mux := sync.RWMutex{}
	r := resty.New()
	var trackers *strset.Set

	updateTrackers := func() error {
		logger.Info("update latest trackers")
		v, err := getTrackers(r)
		if err != nil {
			logger.WithE(err).Error("failed to get latest trackers")

			return err
		}
		mux.Lock()
		trackers = v
		mux.Unlock()
		logger.Info("latest trackers updated")

		return nil
	}

	if err := retry.Do(updateTrackers, retry.Attempts(5), retry.Delay(time.Second)); err != nil {
		logger.WithE(err).Error("failed to update trackers after retries")
	}

	if _, err := c.AddFunc("0 * * * *", func() {
		if err := retry.Do(updateTrackers, retry.Attempts(5), retry.Delay(time.Second)); err != nil {
			logger.WithE(err).Error("failed to update trackers after retries")
		}
	}); err != nil {
		return errgo.Wrap(err, "adding tracker updater")
	}

	_, err := c.AddFunc("* * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		torrents, err := rpc.TorrentGet(ctx,
			[]string{"id", "downloadDir", "labels", "name", "hashString", "trackers"}, nil)
		if err != nil {
			logger.WithE(err).Error("failed to get torrent list")

			return
		}

		for _, torrent := range torrents {
			if err := processTracker(rpc, trackers, &mux, torrent); err != nil {
				logger.WithE(err).Error("failed to update tracker", zap.Stringp("name", torrent.Name))
			}

			if err := processLabels(rpc, torrent); err != nil {
				logger.WithE(err).Error("failed to update labelConfig", zap.Stringp("name", torrent.Name))
			}
		}
	})

	return errgo.Wrap(err, "transmission")
}

func getTrackers(client *resty.Client) (*strset.Set, error) {
	u, ok := os.LookupEnv("TRACKER_LIST")
	if !ok {
		u = "https://trackerslist.com/all.txt"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.R().SetContext(ctx).Get(u)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to fetch latest tracker list")
	}
	if res.StatusCode() > 300 {
		return nil, errgo.Wrap(err,
			fmt.Sprintf("failed to fetch latest tracker list, http code %d", res.StatusCode()))
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(res.Body()))
	trackers := strset.NewWithSize(200)
	for scanner.Scan() {
		v := scanner.Text()

		if shouldAdd(v) {
			trackers.Add(v)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errgo.Wrap(err, "scan")
	}

	logger.Sugar().Infof("updated trackers with lenth %d", trackers.Size())

	return trackers, nil
}

func shouldAdd(s string) bool {
	if s == "" {
		return false
	}

	if u, err := url.Parse(s); err != nil {
		return false
	} else if u.Scheme == "wss" || u.Scheme == "ws" {
		return false
	}

	if trackerShouldRemove.Has(s) {
		return false
	}

	return true
}

//nolint:gochecknoglobals
var trackerShouldRemove = strset.New(
	// only allow authorized info hash
	"http://bt.beatrice-raws.org/announce",
	"http://nyaa.tracker.wf:7777/announce",
	"http://open.touki.ru/announce.php",
	"http://sukebei.tracker.wf:8888/announce",
	"http://torrent.arjlover.net:2710/announce",
	"http://torrent.resonatingmedia.com:6969/announce",
	"http://torrents.hikarinokiseki.com:6969/announce",
	"http://tracker.gcvchp.com:2710/announce",
	"http://tracker.minglong.org:8080/announce",
	"http://tracker.pussytorrents.org:3000/announce",
	"http://tracker.tasvideos.org:6969/announce",
	"http://www.tribalmixes.com/announce.php",
	"https://torrent.ubuntu.com/announce",
	"udp://anidex.moe:6969/announce",

	// cloudflare access deny
	"http://104.28.16.69/announce",
	"https://tracker.shittyurl.org/announce",
	"https://tracker.nitrix.me/announce",
	"https://tracker.lilithraws.cf/announce",
	"https://tracker.nanoha.org/announce",
	"http://www.xwt-classics.net/announce.php",
	"http://torrentsmd.com:8080/announce",
	// bot verify??
	"https://tracker.parrotsec.org/announce",
	// 404
	"http://baibako.tv/announce",
)
