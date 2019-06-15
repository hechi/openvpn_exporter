package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/rajatvig/openvpn_exporter/parsers"
	"os"
)

type Reader struct {
	statusPath string
	client     parsers.Client
	server     parsers.Server
}

func New(statusPath string, ignoreIndividuals bool) (*Reader, error) {
	return &Reader{
		statusPath: statusPath,
		client:     parsers.NewClient(),
		server:     parsers.NewServer(ignoreIndividuals),
	}, nil
}

// Converts OpenVPN status information into Prometheus metrics. This
// function automatically detects whether the file contains server or
// client metrics. For server metrics, it also distinguishes between the
// version 2 and 3 file formats.
func (r *Reader) CollectStatus(ch chan<- prometheus.Metric) error {
	conn, err := os.Open(r.statusPath)
	defer conn.Close()
	if err != nil {
		log.Error("error opening file", err)
		return err
	}

	reader := bufio.NewReader(conn)
	buf, err := reader.Peek(18)
	if err != nil {
		log.Error("error reading file", err)
		return err
	}

	if bytes.HasPrefix(buf, []byte("TITLE,")) {
		// Server statistics, using format version 2.
		return r.server.CollectServerStatusFromReader(r.statusPath, reader, ch, ",")
	} else if bytes.HasPrefix(buf, []byte("TITLE\t")) {
		// Server statistics, using format version 3. The only
		// difference compared to version 2 is that it uses tabs
		// instead of spaces.
		return r.server.CollectServerStatusFromReader(r.statusPath, reader, ch, "\t")
	} else if bytes.HasPrefix(buf, []byte("OpenVPN STATISTICS")) {
		// Client statistics.
		return r.client.CollectClientStatusFromReader(r.statusPath, reader, ch)
	} else {
		return fmt.Errorf("unexpected file contents: %q", buf)
	}
}
