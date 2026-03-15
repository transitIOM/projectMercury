package models

import "time"

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
