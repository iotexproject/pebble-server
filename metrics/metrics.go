package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	deviceRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mqtt_device_requests_total",
			Help: "Total number of mqtt device requests.",
		},
		[]string{"device_id"},
	)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mqtt_requests_total",
			Help: "Total number of mqtt requests.",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(deviceRequestsTotal)
	prometheus.MustRegister(httpRequestsTotal)
}

func TrackDeviceCount(deviceID string) {
	deviceRequestsTotal.WithLabelValues(deviceID).Inc()
}

func TrackRequestCount(method string) {
	httpRequestsTotal.WithLabelValues(method).Inc()
}
