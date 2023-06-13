package utils

import "github.com/prometheus/client_golang/prometheus"

func Gauge(name string, labels prometheus.Labels, value float64) prometheus.Gauge {
	c := prometheus.NewGauge(prometheus.GaugeOpts{Name: name, ConstLabels: labels})
	c.Set(value)
	return c
}
