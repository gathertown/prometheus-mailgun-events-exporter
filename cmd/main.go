package main

import (
	"mailgun_events_exporter/internal/config"
	"mailgun_events_exporter/internal/exporter"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.FromEnv()
	exporter := exporter.NewExporter(cfg)
	exporter.Start()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
