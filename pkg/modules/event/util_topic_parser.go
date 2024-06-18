package event

import "bytes"

type TopicParser struct {
	src    any
	topic  []byte
	prefix string
	suffix string
}

func (m *TopicParser) Unmarshal() error {
	parts := bytes.Split(m.topic, []byte("/"))
	if len(parts) != 3 {
		return &UnmarshalTopicError{topic: string(m.topic), event: m.src}
	}
	if !bytes.Equal(parts[0], []byte(m.prefix)) || !bytes.Equal(parts[2], []byte(m.suffix)) {
		return &UnmarshalTopicError{topic: string(m.topic), event: m.src}
	}
	if len(parts[1]) == 0 {
		return &UnmarshalTopicError{topic: string(m.topic), event: m.src}
	}
	if setter, ok := m.src.(WithIMEI); ok {
		setter.SetIMEI(string(parts[1]))
	}
	return nil
}
