package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type connectionStats struct {
	established uint
	failed      uint
}

type prometheusExporter struct {
	certTTL *prometheus.Desc
	connUp  *prometheus.Desc
	conns   *prometheus.Desc
	fwds    *prometheus.Desc

	connections connectionStats
	forwardings connectionStats
}

var (
	connLabels = []string{"state"}
	hostLabel  = []string{"host"}
)

var metrics = prometheusExporter{
	certTTL: prometheus.NewDesc("sshproxy_certificate_ttl", "TTL until SSH certificate expires", hostLabel, nil),
	connUp:  prometheus.NewDesc("sshproxy_connection_up", "SSH connection up", hostLabel, nil),
	conns:   prometheus.NewDesc("sshproxy_connections_total", "SSH connections", connLabels, nil),
	fwds:    prometheus.NewDesc("sshproxy_forwardings_total", "TCP forwardings", connLabels, nil),
}

// Describe implements (part of the) prometheus.Collector interface.
func (e *prometheusExporter) Describe(c chan<- *prometheus.Desc) {
	c <- metrics.certTTL
	c <- metrics.connUp
	c <- metrics.conns
	c <- metrics.fwds
}

// Collect implements (part of the) prometheus.Collector interface.
func (e prometheusExporter) Collect(c chan<- prometheus.Metric) {
	const C = prometheus.CounterValue
	const G = prometheus.GaugeValue
	met := prometheus.MustNewConstMetric

	c <- met(metrics.conns, C, float64(e.connections.established), "established")
	c <- met(metrics.conns, C, float64(e.connections.failed), "failed")
	c <- met(metrics.fwds, C, float64(e.forwardings.established), "established")
	c <- met(metrics.fwds, C, float64(e.forwardings.failed), "failed")

	proxy.mtx.Lock()
	for key, client := range proxy.clients {
		host := key.String()

		var up float64
		if client.sshClient != nil {
			up = 1
		}
		c <- met(metrics.connUp, G, up, host)

		if cert := client.sshCert; cert != nil {
			ttl := float64(cert.ValidBefore)
			c <- met(metrics.certTTL, G, ttl, host)
		}
	}
	proxy.mtx.Unlock()
}
