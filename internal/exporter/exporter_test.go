package exporter

import (
	"crypto/rand"
	"mailgun_events_exporter/internal/config"
	"mailgun_events_exporter/pkg/mailgun"
	"testing"
	"time"

	"github.com/mailgun/mailgun-go/v4/events"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

type Event struct {
	Accepted  []*events.Accepted
	Delivered []*events.Delivered
	Failed    []*events.Failed
	Err       error
}

type MockMailgunRetriever struct {
	eventSequence []Event
	iteration     int
}

// Given time.Time{} return a float64 as given in mailgun event timestamps
func TimeToFloat(t time.Time) float64 {
	return float64(t.Unix()) + (float64(t.Nanosecond()/int(time.Microsecond)) / float64(1000000))
}

// randomString generates a string of given length, but random content.
// All content will be within the ASCII graphic character set.
// (Implementation from Even Shaw's contribution on
// http://stackoverflow.com/questions/12771930/what-is-the-fastest-way-to-generate-a-long-random-string-in-go).
func randomString(n int, prefix string) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return prefix + string(bytes)
}

func (m *MockMailgunRetriever) GetEvents() ([]*events.Accepted, []*events.Delivered, []*events.Failed, error) {
	if m.iteration >= len(m.eventSequence) {
		return nil, nil, nil, nil
	}

	e := m.eventSequence[m.iteration]
	m.iteration += 1

	return e.Accepted, e.Delivered, e.Failed, e.Err
}

func setupMockMailgunRetriever() mailgun.EventRetriever {
	var (
		recipients      = []string{"one@mailgun.test", "two@mailgun.test"}
		recipientDomain = "mailgun.test"
		timeStamp       = TimeToFloat(time.Now().UTC())
	)

	accepted := new(events.Accepted)
	accepted.ID = randomString(16, "ID-")
	accepted.Message.Headers.MessageID = accepted.ID
	accepted.Name = events.EventAccepted
	accepted.Timestamp = timeStamp
	accepted.Recipient = recipients[0]
	accepted.RecipientDomain = recipientDomain

	delivered := new(events.Delivered)
	delivered.ID = accepted.ID
	delivered.Message.Headers.MessageID = accepted.ID
	delivered.Timestamp = timeStamp + 10
	delivered.Recipient = accepted.Recipient
	delivered.RecipientDomain = accepted.RecipientDomain

	return &MockMailgunRetriever{
		iteration: 0,
		eventSequence: []Event{
			{
				Accepted:  []*events.Accepted{accepted},
				Delivered: nil,
				Failed:    nil,
				Err:       nil,
			},
			{
				Accepted:  nil,
				Delivered: []*events.Delivered{delivered},
				Failed:    []*events.Failed{{}},
				Err:       nil,
			},
		},
	}
}

func TestExporter_RecordMetrics(t *testing.T) {
	cfg := &config.Config{
		Domain:   "example.com",
		ApiKey:   "your-api-key",
		LogLevel: "debug",
	}

	r := setupMockMailgunRetriever()
	e := NewExporter(cfg,
		WithEventRetriever(r),
		WithTicker(time.NewTicker(10*time.Millisecond)))

	e.Start()

	time.Sleep(100 * time.Millisecond)

	// Retrieve and check the registered metrics
	metrics, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	for _, metric := range metrics {
		m := metric.GetMetric()

		if metric.GetName() == "mailgun_delivery_error" {
			assert.Len(t, m, 1)
			assert.Equal(t, float64(1), m[0].GetCounter().GetValue())
		} else if metric.GetName() == "mailgun_delivery_time_seconds" {
			assert.Len(t, m, 1)
			histogram := m[0].GetHistogram()
			assert.Equal(t, uint64(1), histogram.GetSampleCount())
			assert.Equal(t, float64(10), histogram.GetSampleSum())
		}
	}
}
