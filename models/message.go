package models

import (
	"ipfs-sharing/gen/model"
	"time"
)

type MessageStatus string

const (
	MessageStatusSent ContentStatus = "sent"
	MessageStatusRead               = "read"
)

func NewMessage(FromID string, ToID string, Text string) model.Messages {
	now := time.Now().UTC().String()
	return model.Messages{Text: Text, From: FromID, To: ToID, CreatedAt: now}
}
