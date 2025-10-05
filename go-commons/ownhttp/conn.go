package ownhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

type ConnectionOptions struct {
	Action   string   `json:"action"`
	Subjects []string `json:"subjects"`
}

type SocketConnection struct {
	ID    string
	Hub   *Hub
	Conn  *websocket.Conn
	Mutex *sync.Mutex
	done  chan struct{}
	once  sync.Once
	Subs  map[string]struct{}
}

func NewSocketConnection(c *websocket.Conn) *SocketConnection {
	return &SocketConnection{
		Conn:  c,
		ID:    ksuid.New().String(),
		Mutex: &sync.Mutex{},
		done:  make(chan struct{}),
		Subs:  make(map[string]struct{}),
	}
}

func (s *SocketConnection) Close() {
	s.once.Do(func() {
		s.Hub.mu.Lock()
		delete(s.Hub.Clients, s.ID)
		s.Hub.removeAllSubsUnsafe(s)
		s.Hub.mu.Unlock()

		_ = s.Conn.Close()
		close(s.done)
		logrus.WithField("id", s.ID).Info("connection closed")
	})
}

func (s *SocketConnection) Init() {
	s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	msg := fmt.Sprintf(`{"event":"init","data":{"connectionId":"%s"}}`, s.ID)
	_ = s.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s *SocketConnection) Config() {
	s.Conn.SetReadLimit(1 << 20) // 1 MiB
	s.Conn.SetReadDeadline(time.Now().Add(pongWait))

	s.Conn.SetPongHandler(func(string) error {
		s.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	s.Conn.SetPingHandler(func(appData string) error {
		s.Conn.SetReadDeadline(time.Now().Add(pongWait))
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		return s.Conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(writeWait))
	})

	s.Conn.SetCloseHandler(func(code int, text string) error {
		logrus.WithFields(logrus.Fields{"id": s.ID, "code": code, "text": text}).Info("ws close received")
		go s.Close()
		return nil
	})
}

func (s *SocketConnection) Read() {
	defer s.Close()
	s.Config()

	remoteAddr := s.Conn.UnderlyingConn().RemoteAddr()

	for {
		s.Conn.SetReadDeadline(time.Now().Add(pongWait))

		mt, r, err := s.Conn.NextReader()
		if err != nil {
			if ce, ok := err.(*websocket.CloseError); ok {
				logrus.WithFields(logrus.Fields{
					"id":         s.ID,
					"remoteAddr": remoteAddr,
					"code":       ce.Code,
					"text":       ce.Text,
				}).Info("read loop closed by peer")
			} else {
				logrus.WithField("remoteAddr", remoteAddr).
					WithError(err).Info("read loop terminated")
			}
			return
		}

		if mt != websocket.TextMessage {
			_, _ = io.Copy(io.Discard, r)
			continue
		}

		if !s.Hub.IsSubjectMode() {
			_, _ = io.Copy(io.Discard, r)
			continue
		}

		lr := &io.LimitedReader{R: r, N: 1<<20 + 1} // 1 MiB + 1 for overflow

		var opts ConnectionOptions
		dec := json.NewDecoder(lr)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&opts); err != nil {
			logrus.WithError(err).Warn("invalid subscription frame")
			_, _ = io.Copy(io.Discard, lr)
			continue
		}
		if lr.N <= 0 {
			logrus.Warn("frame exceeds limit")
			_, _ = io.Copy(io.Discard, lr)
			continue
		}

		switch strings.ToLower(strings.TrimSpace(opts.Action)) {
		case "subscribe":
			s.Hub.subscribe(s, opts.Subjects)
		case "unsubscribe":
			s.Hub.unsubscribe(s, opts.Subjects)
		default:
			s.MainLog().WithFields(logrus.Fields{
				"action": opts.Action,
				"subs":   opts.Subjects,
			}).Error("unable to take hub action")
		}

		_, _ = io.Copy(io.Discard, lr)
		s.Conn.SetReadDeadline(time.Now().Add(pongWait))
	}
}

func (s *SocketConnection) StartPinger() {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-t.C:
			s.Mutex.Lock()
			s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := s.Conn.WriteMessage(websocket.PingMessage, nil)
			s.Mutex.Unlock()
			if err != nil {
				logrus.WithError(err).Warn("ping failed")
				s.Close()
				return
			}
		}
	}
}

func (s *SocketConnection) MainLog() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"id":         s.ID,
		"remoteAddr": s.Conn.UnderlyingConn().RemoteAddr().String(),
		"subjects":   s.Subs,
	})
}
