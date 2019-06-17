package parsers

import (
	"bufio"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"io"
	"strconv"
	"strings"
)

var (
	// Metrics specific to OpenVPN servers.
	openvpnConnectedClientsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("openvpn", "", "openvpn_server_connected_clients"),
		"Number Of Connected Clients",
		[]string{"name"}, nil)
)

type OpenvpnServerHeader struct {
	LabelColumns []string
	Metrics      []OpenvpnServerHeaderField
}

type OpenvpnServerHeaderField struct {
	Column    string
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

type Server struct {
	openvpnServerHeaders        map[string]OpenvpnServerHeader
}

func NewServer(ignoreIndividuals bool) Server {
	var serverHeaderClientLabels []string
	var serverHeaderClientLabelColumns []string
	var serverHeaderRoutingLabels []string
	var serverHeaderRoutingLabelColumns []string
	if ignoreIndividuals {
		serverHeaderClientLabels = []string{"name", "common_name"}
		serverHeaderClientLabelColumns = []string{"Common Name"}
		serverHeaderRoutingLabels = []string{"name", "common_name"}
		serverHeaderRoutingLabelColumns = []string{"Common Name"}
	} else {
		serverHeaderClientLabels = []string{"name", "common_name", "connection_time", "real_address", "virtual_address", "username"}
		serverHeaderClientLabelColumns = []string{"Common Name", "Connected Since (time_t)", "Real Address", "Virtual Address", "Username"}
		serverHeaderRoutingLabels = []string{"name", "common_name", "real_address", "virtual_address"}
		serverHeaderRoutingLabelColumns = []string{"Common Name", "Real Address", "Virtual Address"}
	}

	openvpnServerHeaders := map[string]OpenvpnServerHeader{
		"CLIENT_LIST": {
			LabelColumns: serverHeaderClientLabelColumns,
			Metrics: []OpenvpnServerHeaderField{
				{
					Column: "Bytes Received",
					Desc: prometheus.NewDesc(
						prometheus.BuildFQName("openvpn", "server", "client_received_bytes_total"),
						"Amount of data received over a connection on the VPN server, in bytes.",
						serverHeaderClientLabels, nil),
					ValueType: prometheus.CounterValue,
				},
				{
					Column: "Bytes Sent",
					Desc: prometheus.NewDesc(
						prometheus.BuildFQName("openvpn", "server", "client_sent_bytes_total"),
						"Amount of data sent over a connection on the VPN server, in bytes.",
						serverHeaderClientLabels, nil),
					ValueType: prometheus.CounterValue,
				},
			},
		},
		"ROUTING_TABLE": {
			LabelColumns: serverHeaderRoutingLabelColumns,
			Metrics: []OpenvpnServerHeaderField{
				{
					Column: "Last Ref (time_t)",
					Desc: prometheus.NewDesc(
						prometheus.BuildFQName("openvpn", "server", "route_last_reference_time_seconds"),
						"Time at which a route was last referenced, in seconds.",
						serverHeaderRoutingLabels, nil),
					ValueType: prometheus.GaugeValue,
				},
			},
		},
	}

	return Server{
		openvpnServerHeaders:        openvpnServerHeaders,
	}
}

// Converts OpenVPN server status information into Prometheus metrics.
func (s Server) CollectServerStatusFromReader(name string, file io.Reader, ch chan<- prometheus.Metric, separator string) error {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	headersFound := map[string][]string{}
	// counter of connected client
	numberConnectedClient := 0

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), separator)
		if fields[0] == "END" && len(fields) == 1 {
			// Stats footer.
		} else if fields[0] == "GLOBAL_STATS" {
			// Global server statistics.
		} else if fields[0] == "HEADER" && len(fields) > 2 {
			// Column names for CLIENT_LIST and ROUTING_TABLE.
			headersFound[fields[1]] = fields[2:]
		} else if fields[0] == "TIME" && len(fields) == 3 {
			// Time at which the statistics were updated.
			timeStartStats, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				log.Error("error parsing float", err)
				return err
			}
			ch <- prometheus.MustNewConstMetric(
				openvpnStatusUpdateTimeDesc,
				prometheus.GaugeValue,
				timeStartStats,
				name)
		} else if fields[0] == "TITLE" && len(fields) == 2 {
			// OpenVPN version number.
		} else if header, ok := s.openvpnServerHeaders[fields[0]]; ok {
			if fields[0] == "CLIENT_LIST" {
				numberConnectedClient++
			}
			// Entry that depends on a preceding HEADERS directive.
			columnNames, ok := headersFound[fields[0]]
			if !ok {
				return fmt.Errorf("%s should be preceded by HEADERS", fields[0])
			}
			if len(fields) != len(columnNames)+1 {
				return fmt.Errorf("HEADER for %s describes a different number of columns", fields[0])
			}

			// Store entry values in a map indexed by column name.
			columnValues := map[string]string{}
			for _, column := range header.LabelColumns {
				columnValues[column] = ""
			}
			for i, column := range columnNames {
				columnValues[column] = fields[i+1]
			}

			// Extract columns that should act as entry labels.
			labels := []string{name}
			for _, column := range header.LabelColumns {
				labels = append(labels, columnValues[column])
			}

			// Export relevant columns as individual metrics.
			for _, metric := range header.Metrics {
				if columnValue, ok := columnValues[metric.Column]; ok {
					value, err := strconv.ParseFloat(columnValue, 64)
					if err != nil {
						log.Error("error parsing float", err)
						return err
					}
					ch <- prometheus.MustNewConstMetric(
						metric.Desc,
						metric.ValueType,
						value,
						labels...)
				}
			}
		} else {
			return fmt.Errorf("unsupported key: %q", fields[0])
		}
	}
	// add the number of connected client
	ch <- prometheus.MustNewConstMetric(
		openvpnConnectedClientsDesc,
		prometheus.GaugeValue,
		float64(numberConnectedClient),
		name)

	return scanner.Err()
}
