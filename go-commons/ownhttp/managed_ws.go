package ownhttp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ManagedWS struct {
	Name      string
	Url       string
	conn      *websocket.Conn
	healthy   atomic.Bool
	mu        sync.Mutex
	Messages  chan []byte
	OnConnect func(*websocket.Conn) error
	OnStatus  func(name, status string)
}

func (m *ManagedWS) Start(ctx context.Context) {
	go m.readerLoop(ctx)
}

func (m *ManagedWS) IsHealthy() bool {
	return m.healthy.Load()
}

func (m *ManagedWS) Close() {
	if m.conn != nil {
		_ = m.conn.Close()
	}
}

func (m *ManagedWS) SendJSON(v interface{}) error {
	if m.conn == nil {
		return fmt.Errorf("ws %s not connected", m.Name)
	}
	return m.conn.WriteJSON(v)
}

func (m *ManagedWS) readerLoop(ctx context.Context) {
	backoff := time.Second

	for {
		if m.conn == nil {
			if err := m.reconnect(ctx); err != nil {
				logrus.WithError(err).Errorf("failed to connect ws %s", m.Name)
				time.Sleep(backoff)
				if backoff < 30*time.Second {
					backoff *= 2
				}
				continue
			}
			backoff = time.Second
		}

		_, data, err := m.conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Warnf("ws %s closed, reconnecting...", m.Name)
			m.healthy.Store(false)
			if m.OnStatus != nil {
				m.OnStatus(m.Name, "DOWN")
			}
			m.conn = nil
			continue
		}

		if !m.healthy.Load() {
			m.healthy.Store(true)
			if m.OnStatus != nil {
				m.OnStatus(m.Name, "RECOVERED")
			}
		}

		select {
		case m.Messages <- data:
		case <-ctx.Done():
			return
		}
	}
}

func (m *ManagedWS) reconnect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		_ = m.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "reconnect"))
		_ = m.conn.Close()
		m.conn = nil
	}

	client, err := NewWSClient(ctx, m.Url)
	if err != nil {
		return err
	}

	m.conn = client.Conn
	if m.OnConnect != nil {
		if err := m.OnConnect(m.conn); err != nil {
			return err
		}
	}

	logrus.Infof("ws %v connected and resubscribed url=%s", m.Name, m.Url)
	return nil
}
