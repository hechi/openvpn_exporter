package parsers

import "github.com/prometheus/client_golang/prometheus"

var (
	openvpnStatusUpdateTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "", "status_update_time_seconds"),
		"UNIX timestamp at which the OpenVPN statistics were updated.",
		[]string{"name"}, nil)
)
