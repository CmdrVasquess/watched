package edj

import "time"

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
}
