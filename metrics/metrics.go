package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	deviceRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "device_requests_total",
			Help: "Total number of device requests.",
		},
		[]string{"device_id"},
	)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method"},
	)
	httpDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_duration",
			Help:    "Histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(deviceRequestsTotal)
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpDurationHistogram)
}

func TrackDeviceCount(deviceID string) {
	deviceRequestsTotal.WithLabelValues(deviceID).Inc()
}

func TrackRequestCount(method string) {
	httpRequestsTotal.WithLabelValues(method).Inc()
}

func TrackRequestDuration(method string, duration time.Duration) {
	httpDurationHistogram.WithLabelValues(method).Observe(float64(duration))
}
