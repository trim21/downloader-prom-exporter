package reporter

import (
	"app/pkg/reporter/qbittorrent"
	"app/pkg/reporter/rtorrent"
	"app/pkg/reporter/transmission"
)

func SetupMetrics() error {
	if err := transmission.SetupMetrics(); err != nil {
		return err
	}

	if err := qbittorrent.SetupMetrics(); err != nil {
		return err
	}

	if err := rtorrent.SetupMetrics(); err != nil {
		return err
	}

	return nil
}
