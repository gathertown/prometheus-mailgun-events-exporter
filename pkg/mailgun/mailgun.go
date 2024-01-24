package mailgun

import (
	"context"
	"fmt"
	"mailgun_events_exporter/internal/config"
	"mailgun_events_exporter/pkg/log"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/mailgun/mailgun-go/v4/events"
)

var cfg = config.FromEnv()
var logger = log.New(os.Stdout, cfg.LogLevel)

type EventRetriever interface {
	GetEvents() ([]*events.Accepted, []*events.Delivered, []*events.Failed, error)
}

type MailgunRetriever struct {
	EventRetriever

	mg *mailgun.MailgunImpl
}

func NewMailgunRetriever(domain, apiKey string) *MailgunRetriever {
	return &MailgunRetriever{
		mg: mailgun.NewMailgun(domain, apiKey),
	}
}

func (r *MailgunRetriever) GetEvents() ([]*events.Accepted, []*events.Delivered, []*events.Failed, error) {
	accepted := []*events.Accepted{}
	delivered := []*events.Delivered{}
	failed := []*events.Failed{}

	it := r.mg.ListEvents(&mailgun.ListEventOptions{
		Begin: time.Now().Add(-3 * time.Minute),
		End:   time.Now().Add(-2 * time.Minute),
		Limit: 300,
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
