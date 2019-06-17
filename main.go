// Copyright 2017 Kumina, https://kumina.nl/
// Copyright 2019 Rajat Vig, https://rajatvig.keybase.pub/
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/rajatvig/openvpn_exporter/collector"
	"github.com/rajatvig/openvpn_exporter/config"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").
		Default(":9176").
		String()
	metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").
		Default("/metrics").
		String()
	configFile = kingpin.Flag("config.file", "Config file for the exporter").
		Default("examples/config.yaml").
		String()
	ignoreIndividuals = kingpin.Flag("ignore.individuals", "If ignoring metrics for individuals").
		Default("true").
		Bool()
)

func main() {
	os.Exit(run())
}

func run() int {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("prom-metrics-writer"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Info("Starting OpenVPN Exporter\n")
	log.Infof("Listen address: %v\n", *listenAddress)
	log.Infof("Metrics path: %v\n", *metricsPath)
	log.Infof("Configuration File: %v\n", *configFile)
	log.Infof("Ignore Individuals: %v\n", *ignoreIndividuals)

	sc := config.SafeConfig{}
	err := sc.Load(*configFile)
	if err != nil {
		log.Error("error reading config", err)
		return 1
	}

	exporter := collector.OpenVpn{
		Configs:           sc.C.Config,
		IgnoreIndividuals: *ignoreIndividuals,
	}

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>OpenVPN Exporter</title></head>
			<body>
			<h1>OpenVPN Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	srv := http.Server{Addr: *listenAddress}
	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("Listening on address", "address", *listenAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	for {
		select {
		case <-term:
			log.Info("Received SIGTERM, exiting gracefully...")
			return 0
		case <-srvc:
			return 1
		}
	}
}
