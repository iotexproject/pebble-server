package event

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xoctopus/x/misc/stringsx"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

var (
	mtcTotalEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_event",
			Help: "total event count",
		},
		[]string{"source_type", "topic"},
	)
	mtcFailedEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_failed_event",
			Help: "total failed event count",
		},
		[]string{"source_type", "topic"},
	)
	mtcSucceededEvent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_succeeded_event",
			Help: "total succeeded event count",
		},
		[]string{"source_type", "topic"},
	)
	mtcEventHandlingCost = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "event_handling_cost",
			Help: "event handling cost",
		},
		[]string{"source_type", "topic"},
	)
)

func init() {
	prometheus.MustRegister(mtcTotalEvent)
	prometheus.MustRegister(mtcFailedEvent)
	prometheus.MustRegister(mtcSucceededEvent)
	prometheus.MustRegister(mtcEventHandlingCost)
}

func stat(v Event, e error, cost time.Duration) {
	var (
		source = v.Source().String()
		topic  = v.Topic()
	)
	if source == enums.EVENT_SOURCE_TYPE__MQTT.String() {
		topic = stringsx.UpperSnakeCase(topic)
	}

	mtcTotalEvent.WithLabelValues(source, topic).Inc()
	if e != nil {
		mtcFailedEvent.WithLabelValues(source, topic).Inc()
	} else {
		mtcSucceededEvent.WithLabelValues(source, topic).Inc()
	}
	mtcEventHandlingCost.WithLabelValues(source, topic).Observe(float64(cost.Milliseconds()))
}
