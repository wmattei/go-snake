package communication

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type CommunicationLayer interface {
	Initialize() error
	Connect() error
	SendMessage(message Message) error
	Listen(listener chan Message) error
	Close() error
}
