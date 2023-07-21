package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func Gauge(name string, labels prometheus.Labels, value float64, help string) prometheus.Metric {
	return gauge{name, labels, value, help}
}

type gauge struct {
	name   string
	labels prometheus.Labels
	value  float64
	help   string
}

func (c gauge) Desc() *prometheus.Desc {
	return prometheus.NewDesc(c.name, c.help, nil, nil)
}

func (c gauge) Write(metric *dto.Metric) error {
	metric.Gauge = &dto.Gauge{Value: &c.value}
	metric.Label = toLabelParis(c.labels)

	return nil
}

func Count(name string, labels prometheus.Labels, value float64, help string) prometheus.Metric {
	return count{name, labels, value, help}
}

type count struct {
	name   string
	labels prometheus.Labels
	value  float64
	help   string
}

func (c count) Desc() *prometheus.Desc {
	return prometheus.NewDesc(c.name, c.help, nil, c.labels)
}

func (c count) Write(metric *dto.Metric) error {
	metric.Counter = &dto.Counter{Value: &c.value}
	metric.Label = toLabelParis(c.labels)

	return nil
}

func toLabelParis(labels prometheus.Labels) []*dto.LabelPair {
	var r = make([]*dto.LabelPair, 0, len(labels))
	for key, value := range labels {
		key := key
		value := value
		r = append(r, &dto.LabelPair{
			Name:  &key,
			Value: &value,
		})
	}

	return r
}
