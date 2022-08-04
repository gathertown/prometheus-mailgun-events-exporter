package main

import (
	"mailgun_events_exporter/internal/config"
	"mailgun_events_exporter/pkg/log"
	"mailgun_events_exporter/pkg/mailgun"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cfg = config.FromEnv()
var logger = log.New(os.Stdout, cfg.LogLevel)

func main() {

	go recordMetrics(cfg)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func recordMetrics(cfg *config.Config) {
	ticker := time.NewTicker(time.Minute)
	for ; true; <-ticker.C {
		logger.Debug("")
		accepted, delivered, failed, err := mailgun.GetMailgunEventsPerType(cfg.Domain, cfg.ApiKey)
		if err != nil {
			logger.Error(err.Error())
		}
		mailgun.RecordDeliverySpeed(accepted, delivered)
		mailgun.RecordDeliveryErrorMessages(failed)
		logger.Debug("")
	}

}
