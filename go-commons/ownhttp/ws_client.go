package ownhttp

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WSClient struct {
	Conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWSClient(ctx context.Context, url string) (*WSClient, error) {

	headers := HeadersFromEnv()

	c, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		return nil, err
	}

	childCtx, cancel := context.WithCancel(ctx)
	client := &WSClient{Conn: c, ctx: childCtx, cancel: cancel}

	// keepalive ping cada 30s
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-childCtx.Done():
				logrus.Info("Finalizing ws client")
				return
			case <-ticker.C:
				_ = c.WriteMessage(websocket.PingMessage, []byte("ping"))
			}
		}
	}()

	return client, nil
}

func (w *WSClient) Close() {
	w.cancel()
	if w.Conn != nil {
		_ = w.Conn.Close()
	}
}
