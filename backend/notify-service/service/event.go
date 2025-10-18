package service

import "github.com/nats-io/nats.go"

type NotifyEvent struct {
	Subject string      `json:"subject"`
	ID      string      `json:"id"`
	Data    []byte      `json:"data"`
	Header  nats.Header `json:"header"`
}
