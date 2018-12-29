package main

import "github.com/prometheus/client_golang/prometheus"

type connectionStats struct {
	established uint
	failed      uint
}

type prometheusExporter struct {
	connections connectionStats
	forwardings connectionStats
}

var (
	metrics = prometheusExporter{}

	sshConnEstablishedDesc    = prometheus.NewDesc("sshproxy_connections_established", "Established SSH connections", nil, nil)
	sshConnFailedDesc         = prometheus.NewDesc("sshproxy_connections_failed", "Failed SSH connections", nil, nil)
	sshForwardEstablishedDesc = prometheus.NewDesc("sshproxy_forwardings_established", "Established SSH forwardings", nil, nil)
	sshForwardFailedDesc      = prometheus.NewDesc("sshproxy_forwardings_failed", "Failed SSH forwardings", nil, nil)
)

// Describe implements (part of the) prometheus.Collector interface.
func (e *prometheusExporter) Describe(c chan<- *prometheus.Desc) {
	c <- sshConnEstablishedDesc
	c <- sshConnFailedDesc
	c <- sshForwardEstablishedDesc
	c <- sshForwardFailedDesc
}

// Collect implements (part of the) prometheus.Collector interface.
func (e prometheusExporter) Collect(c chan<- prometheus.Metric) {
	c <- prometheus.MustNewConstMetric(sshConnEstablishedDesc, prometheus.CounterValue, float64(e.connections.established))
	c <- prometheus.MustNewConstMetric(sshConnFailedDesc, prometheus.CounterValue, float64(e.connections.failed))
	c <- prometheus.MustNewConstMetric(sshForwardEstablishedDesc, prometheus.CounterValue, float64(e.forwardings.established))
	c <- prometheus.MustNewConstMetric(sshForwardFailedDesc, prometheus.CounterValue, float64(e.forwardings.failed))
}
