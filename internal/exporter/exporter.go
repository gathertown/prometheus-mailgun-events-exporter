package exporter

import (
	"fmt"
	"mailgun_events_exporter/internal/config"
	"mailgun_events_exporter/internal/metrics"
	"mailgun_events_exporter/pkg/log"
	"mailgun_events_exporter/pkg/mailgun"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4/events"
	"github.com/zekroTJA/timedmap"
)

type Exporter struct {
	started bool
	cfg     *config.Config
	ticker  *time.Ticker
	tm      *timedmap.TimedMap
	logger  *log.Logger
	mg      mailgun.EventRetriever
}

type ExporterOption func(*Exporter)

const (
	DEFAULT_RETENTION_DURATION = 10 * time.Hour
)

func messageKey(messageId string, recipient string) string {
	return fmt.Sprintf("%s/%s", messageId, recipient)
}

func WithEventRetriever(mg mailgun.EventRetriever) ExporterOption {
	return func(e *Exporter) {
		e.mg = mg
	}
}

func WithTicker(ticker *time.Ticker) ExporterOption {
	return func(e *Exporter) {
		e.ticker = ticker
	}
}

func NewExporter(cfg *config.Config, opts ...ExporterOption) *Exporter {
	e := &Exporter{
		started: false,
		cfg:     cfg,
		ticker:  time.NewTicker(time.Minute),
		tm:      timedmap.New(10 * time.Minute),
		logger:  log.New(os.Stdout, cfg.LogLevel),
		mg:      mailgun.NewMailgunRetriever(cfg.Domain, cfg.ApiKey),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *Exporter) Start() {
	if !e.started {
		go e.recordMetrics()
		e.started = true
	}
}

func (e *Exporter) recordMetrics() {
	for ; true; <-e.ticker.C {
		accepted, delivered, failed, err := e.mg.GetEvents()
		if err != nil {
			e.logger.Error(err.Error())
		}

		e.recordAccepted(accepted)
		e.recordDelivered(delivered)
		e.recordFailed(failed)
	}
}

func (e *Exporter) recordAccepted(accepted []*events.Accepted) {
	for _, a := range accepted {
		key := messageKey(a.Message.Headers.MessageID, a.Recipient)
		e.tm.Set(key, a.Timestamp, DEFAULT_RETENTION_DURATION)
	}
}

func (e *Exporter) recordDelivered(delivered []*events.Delivered) {
	for _, d := range delivered {
		key := messageKey(d.Message.Headers.MessageID, d.Recipient)
		acceptedTime, ok := e.tm.GetValue(key).(float64)
		if !ok {
			continue
		}

		e.tm.Remove(key)

		delta := d.Timestamp - acceptedTime
		if delta < 0 {
			e.logger.Error("Delivery time is negative")
			continue
		}

		e.logger.Debug("Delivery Succeeded", "MessageID", d.Message.Headers.MessageID, "Accepted at", acceptedTime, "Delivered at", d.GetTimestamp(), "Delivery time", delta)
		metrics.DeliveryTime.WithLabelValues(e.cfg.Domain).Observe(delta)
	}
}

func (e *Exporter) recordFailed(failed []*events.Failed) {
	for _, f := range failed {
		severity := f.Severity

		if severity == "permanent" {
			key := messageKey(f.Message.Headers.MessageID, f.Recipient)
			e.tm.Remove(key)
		}

		e.logger.Debug("Delivery Failed", "MessageID", f.Message.Headers.MessageID, "Failed at", f.GetTimestamp(), "Reason", f.Reason, "Error Message", f.DeliveryStatus.Message, "Failure Severity", severity)
		metrics.DeliveryError.WithLabelValues(e.cfg.Domain, f.Reason, severity, fmt.Sprintf("%d", f.DeliveryStatus.Code)).Inc()
	}
}
