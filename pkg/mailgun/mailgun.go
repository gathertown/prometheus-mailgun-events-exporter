package mailgun

import (
	"context"
	"fmt"
	"mailgun_events_exporter/internal/config"
	"mailgun_events_exporter/internal/metrics"
	"mailgun_events_exporter/pkg/log"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/mailgun/mailgun-go/v4/events"
)

var cfg = config.FromEnv()
var logger = log.New(os.Stdout, cfg.LogLevel)

func GetMailgunEventsPerType(domain, apiKey string) ([]*events.Accepted, []*events.Delivered, []*events.Failed, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	accepted := []*events.Accepted{}
	delivered := []*events.Delivered{}
	failed := []*events.Failed{}

	it := mg.ListEvents(&mailgun.ListEventOptions{
		Begin: time.Now().Add(-3 * time.Minute),
		End:   time.Now().Add(-2 * time.Minute),
		Limit: 100,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Iterate through all the pages of events
	var page []mailgun.Event
	for it.Next(ctx, &page) {
		for _, event := range page {
			switch e := event.(type) {
			case *events.Accepted:
				accepted = append(accepted, e)
			case *events.Delivered:
				delivered = append(delivered, e)
			case *events.Failed:
				failed = append(failed, e)
			}
		}
	}

	if it.Err() != nil {
		logger.Error(it.Err().Error())
		fmt.Println(it.Err())
		return nil, nil, nil, it.Err()
	}
	logger.Debug("Total Events last minute", "Accepted", len(accepted), "Delivered", len(delivered), "Failed", len(failed))
	return accepted, delivered, failed, nil
}

func RecordDeliverySpeed(accepted []*events.Accepted, delivered []*events.Delivered) {
	for _, a := range accepted {
		for _, d := range delivered {
			if d.Message.Headers.MessageID == a.Message.Headers.MessageID {
				logger.Debug("Delivery Succeeded", "MessageID", a.Message.Headers.MessageID, "Accepted at", a.GetTimestamp(), "Delivered at", d.GetTimestamp(), "Delivery time", d.Timestamp-a.Timestamp)
				metrics.DeliveryTime.WithLabelValues(a.Message.Headers.MessageID).Set(float64(d.Timestamp - a.Timestamp))
				break
			}
		}
	}
}

func RecordDeliveryErrorMessages(failed []*events.Failed) {
	for _, f := range failed {
		logger.Debug("Delivery Failed", "MessageID", f.Message.Headers.MessageID, "Failed at", f.GetTimestamp(), "Reason", f.Reason, "Error Message", f.DeliveryStatus.Message, "Failure Severity", f.Severity)
		metrics.DeliveryError.WithLabelValues(f.Message.Headers.MessageID, f.Reason, f.DeliveryStatus.Message, f.Severity).Inc()
	}
}
