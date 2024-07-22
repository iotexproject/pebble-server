package enums

// SourceType defines event source types
type EventSourceType uint8

const (
	EVENT_SOURCE_TYPE__MQTT EventSourceType = iota + 1
	EVENT_SOURCE_TYPE__BLOCKCHAIN
)

func (v EventSourceType) String() string {
	switch v {
	default:
		return "UNKNOWN"
	case EVENT_SOURCE_TYPE__BLOCKCHAIN:
		return "BLOCKCHAIN"
	case EVENT_SOURCE_TYPE__MQTT:
		return "MQTT"
	}
}
