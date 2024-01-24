package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	namespace     = "mailgun"
	DeliveryError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "delivery_error",
		Namespace: namespace,
		Help:      "Email Delivery errors",
	}, []string{"domain", "reason", "severity", "delivery_status_code"},
	)

	DeliveryTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "delivery_time_seconds",
		Namespace: namespace,
		Help:      "The time took for an email to actually got delivered from the time that got accepted in mailgun",
		Buckets:   []float64{0.5, 1, 2, 5, 10, 20, 40, 60, 120, 300, 600, 1800, 3600},
	}, []string{"domain"},
	)
)
