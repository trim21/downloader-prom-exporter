package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func Gauge(name string, labels prometheus.Labels, value float64) prometheus.Gauge {
	c := prometheus.NewGauge(prometheus.GaugeOpts{Name: name, ConstLabels: labels})
	c.Set(value)
	return c
}

func Count(name string, labels prometheus.Labels, value float64) prometheus.Metric {
	return count{name, labels, value}
}

type count struct {
	name   string
	labels prometheus.Labels
	value  float64
}

func (c count) Desc() *prometheus.Desc {
	return prometheus.NewDesc(c.name, "", nil, c.labels)
}

func (c count) Write(metric *dto.Metric) error {
	metric.Counter = &dto.Counter{Value: &c.value}

	return nil
}
