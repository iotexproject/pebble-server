package event

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xoctopus/x/misc/stringsx"

	"github.com/iotexproject/pebble-server/enums"
	"github.com/iotexproject/pebble-server/models"
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
	mtcEventHandlingCost.WithLabelValues(source, topic).Observe(cost.Seconds())
}

// submit customized metrics to clickhouse(maybe deprecated)
func submit(ctx context.Context, r *models.DeviceRecord) {
	_ = r.Longitude
	_ = r.Latitude
	_ = fmt.Sprintf(`{"longitude":%s,"latitude":%s}`, r.Longitude, r.Latitude)
	// INSERT INTO ws_metrics.auto_collect_metrics VALUES (now(),'%s','%s','%s','%s')
	// account_id, project_name, publish_key, {"longitude": lon,"latitude": lat}
	// INSERT INTO ws_metrics.customized_metrics VALUES (now(), '%s','%s','%s')
	// account_id, project_name,
}
