package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/rajatvig/openvpn_exporter/config"
	"github.com/rajatvig/openvpn_exporter/reader"
)

var (
	// Metrics exported both for client and server statistics.
	openvpnUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "", "up"),
		"Whether scraping OpenVPN's metrics was successful.",
		[]string{"name"}, nil)
)

type OpenVpn struct {
	Configs           []config.Config
	IgnoreIndividuals bool
}

func (o OpenVpn) Describe(ch chan<- *prometheus.Desc) {
	ch <- openvpnUpDesc
}

func (o OpenVpn) Collect(ch chan<- prometheus.Metric) {
	for _, c := range o.Configs {
		r, err := reader.New(c, o.IgnoreIndividuals)
		if err != nil {
			log.Error("failed to create reader", err)
			ch <- prometheus.MustNewConstMetric(openvpnUpDesc, prometheus.GaugeValue, 0.0, c.Name)
			return
		}
		err = r.CollectStatus(ch)
		if err != nil {
			log.Error("failed to scrape showq socket", err)
			ch <- prometheus.MustNewConstMetric(openvpnUpDesc, prometheus.GaugeValue, 0.0, c.Name)
			return
		}

		ch <- prometheus.MustNewConstMetric(openvpnUpDesc, prometheus.GaugeValue, 1.0, c.Name)
	}
}
