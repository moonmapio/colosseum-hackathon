package service

import "github.com/nats-io/nats.go"

type NotifyEvent struct {
	Subject string
	ID      string
	Data    []byte
	Header  nats.Header
}
