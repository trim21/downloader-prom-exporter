package reporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

func SetupMetrics() error {
	if reporter, err := setupTransmissionMetrics(); err != nil {
		return err
	} else if reporter != nil {
		prometheus.MustRegister(reporter)
	}

	if reporter := setupRTorrentMetrics(); reporter != nil {
		prometheus.MustRegister(reporter)
	}

	if reporter := setupQBitMetrics(); reporter != nil {
		prometheus.MustRegister(reporter)
	}

	return nil
}
