package mqtt

type message struct {
	topic   string
	payload []byte
}

func (m *message) Topic() string {
	return m.topic
}

func (m *message) Payload() []byte {
	return m.payload
}
