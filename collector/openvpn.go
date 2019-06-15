package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/rajatvig/openvpn_exporter/reader"
)

var (
	// Metrics exported both for client and server statistics.
	openvpnUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "", "up"),
		"Whether scraping OpenVPN's metrics was successful.",
		[]string{"status_path"}, nil)
)

type OpenVpn struct {
	StatusPaths       []string
	IgnoreIndividuals bool
}

func (o OpenVpn) Describe(ch chan<- *prometheus.Desc) {
	ch <- openvpnUpDesc
}

func (o OpenVpn) Collect(ch chan<- prometheus.Metric) {
	for _, statusPath := range o.StatusPaths {
		r, err := reader.New(statusPath, o.IgnoreIndividuals)
		if err != nil {
			log.Error("failed to create reader", err)
			ch <- prometheus.MustNewConstMetric(openvpnUpDesc, prometheus.GaugeValue, 0.0, statusPath)
			return
		}
		err = r.CollectStatus(ch)
		if err != nil {
			log.Error("failed to scrape showq socket", err)
			ch <- prometheus.MustNewConstMetric(openvpnUpDesc, prometheus.GaugeValue, 0.0, statusPath)
			return
		}

		ch <- prometheus.MustNewConstMetric(openvpnUpDesc, prometheus.GaugeValue, 1.0, statusPath)
	}
}
