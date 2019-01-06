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

	variableLabels     = []string{"state"}
	sshConnectionsDesc = prometheus.NewDesc("sshproxy_connections_total", "SSH connections", variableLabels, nil)
	sshForwardingsDesc = prometheus.NewDesc("sshproxy_forwardings_total", "TCP forwardings", variableLabels, nil)
)

// Describe implements (part of the) prometheus.Collector interface.
func (e *prometheusExporter) Describe(c chan<- *prometheus.Desc) {
	c <- sshConnectionsDesc
	c <- sshForwardingsDesc
}

// Collect implements (part of the) prometheus.Collector interface.
func (e prometheusExporter) Collect(c chan<- prometheus.Metric) {
	c <- prometheus.MustNewConstMetric(sshConnectionsDesc, prometheus.CounterValue, float64(e.connections.established), "established")
	c <- prometheus.MustNewConstMetric(sshConnectionsDesc, prometheus.CounterValue, float64(e.connections.failed), "failed")
	c <- prometheus.MustNewConstMetric(sshForwardingsDesc, prometheus.CounterValue, float64(e.forwardings.established), "established")
	c <- prometheus.MustNewConstMetric(sshForwardingsDesc, prometheus.CounterValue, float64(e.forwardings.failed), "failed")
}
