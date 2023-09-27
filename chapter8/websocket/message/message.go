package message

import "time"

type Message struct {
	Timestamp time.Time
	Message   string
	From      string
}
