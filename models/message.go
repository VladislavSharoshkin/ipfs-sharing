package models

type Message struct {
	FromID string
	ToID   string
	Text   string
}

func NewMessage(FromID string, ToID string, Text string) Message {
	return Message{FromID, ToID, Text}
}
