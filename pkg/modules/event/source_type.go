package event

// SourceType defines event source types
type SourceType uint8

const (
	SOURCE_TYPE__MQTT = iota + 1
	SOURCE_TYPE__BLOCKCHAIN
)

func (v SourceType) String() string {
	switch v {
	default:
		return "UNKNOWN"
	case SOURCE_TYPE__BLOCKCHAIN:
		return "BLOCKCHAIN"
	case SOURCE_TYPE__MQTT:
		return "MQTT"
	}
}
