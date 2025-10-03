// ownhttp/hub.go
package ownhttp

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 25 * time.Second
)

type Hub struct {
	mu       sync.RWMutex
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]*sync.Mutex // write mutex por conexión
	subjects map[string]map[*websocket.Conn]*sync.Mutex
	Mode     string
}

func NewHub() *Hub {
	return &Hub{
		clients:  make(map[*websocket.Conn]*sync.Mutex),
		subjects: make(map[string]map[*websocket.Conn]*sync.Mutex),
		Mode:     "all", // "all" | "subjects"
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

func (h *Hub) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) Stats() map[string]map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := make(map[string]map[string]any, len(h.subjects))
	for subj, conns := range h.subjects {
		ips := make([]string, 0, len(conns))
		for c := range conns {
			ips = append(ips, c.RemoteAddr().String())
		}
		stats[subj] = map[string]any{
			"qty": len(conns),
			"ips": ips,
		}
	}
	return stats
}

func (h *Hub) Add(w http.ResponseWriter, r *http.Request) {
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logrus.WithField("remote_conn", c.UnderlyingConn().RemoteAddr().String()).Info("ws upgraded")

	// registra cliente con su write-mutex
	h.mu.Lock()
	wmu := &sync.Mutex{}
	h.clients[c] = wmu
	h.mu.Unlock()

	// saludo
	wmu.Lock()
	c.SetWriteDeadline(time.Now().Add(writeWait))
	_ = c.WriteMessage(websocket.TextMessage, []byte(`{"event":"hello","data":"connected"}`))
	wmu.Unlock()

	// READER: renueva deadline con PONG y con cualquier frame recibido
	go func(conn *websocket.Conn) {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			for subj := range h.subjects {
				delete(h.subjects[subj], conn)
				if len(h.subjects[subj]) == 0 {
					delete(h.subjects, subj)
				}
			}
			h.mu.Unlock()
			_ = conn.Close()
			logrus.Info("ws reader closed")
		}()
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// heartbeat de app: si el cliente manda "ping", responde pong de control
			if len(msg) == 4 && string(msg) == "ping" {
				wmu.Lock()
				_ = conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(writeWait))
				wmu.Unlock()
				continue
			}

			if strings.EqualFold(h.Mode, "subjects") {
				var frame struct {
					Action   string   `json:"action"`
					Subjects []string `json:"subjects"`
				}
				if err := json.Unmarshal(msg, &frame); err == nil {
					switch frame.Action {
					case "subscribe":
						h.Subscribe(conn, frame.Subjects)
						logrus.WithFields(logrus.Fields{
							"conn":     conn.RemoteAddr().String(),
							"subjects": frame.Subjects,
						}).Info("client subscribed")

					case "unsubscribe":
						if len(frame.Subjects) == 0 {
							// si no pasan subjects, desuscribir de todos
							h.UnsubscribeAll(conn)
							logrus.WithField("conn", conn.RemoteAddr().String()).Info("client unsubscribed from all")
						} else {
							// desuscribir solo de los indicados
							h.Unsubscribe(conn, frame.Subjects)
							logrus.WithFields(logrus.Fields{
								"conn":     conn.RemoteAddr().String(),
								"subjects": frame.Subjects,
							}).Info("client unsubscribed")
						}
					}
				}
			}

			// renueva deadline aunque no haya PONG (útil para clientes que no lo envían, ej. Postman)
			conn.SetReadDeadline(time.Now().Add(pongWait))
		}
	}(c)

	// PINGER: envía Ping cada pingPeriod y extiende deadline proactivamente
	go func(conn *websocket.Conn) {
		t := time.NewTicker(pingPeriod)
		defer func() {
			t.Stop()
			_ = conn.Close()
			logrus.Info("ws writer closed")
		}()
		for range t.C {
			// extiende para clientes que no responden PONG (Postman)
			conn.SetReadDeadline(time.Now().Add(pongWait))
			wmu.Lock()
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				wmu.Unlock()
				return
			}
			wmu.Unlock()
		}
	}(c)
}

func (h *Hub) BroadcastJSON(eventName string, data any) {
	logrus.Debugf("broadcasting %v", eventName)
	payload := map[string]any{"event": eventName, "data": data}
	b, _ := json.Marshal(payload)

	h.mu.RLock()
	var targets map[*websocket.Conn]*sync.Mutex
	if h.Mode == "all" {
		// broadcast a todos
		targets = h.clients
	} else {
		// broadcast solo a subs de ese evento
		targets = h.subjects[eventName]
	}

	// snapshot para no bloquear h.mu
	ps := make([]struct {
		c  *websocket.Conn
		mu *sync.Mutex
	}, 0, len(targets))
	for c, mu := range targets {
		ps = append(ps, struct {
			c  *websocket.Conn
			mu *sync.Mutex
		}{c, mu})
	}
	h.mu.RUnlock()

	// enviar
	for _, p := range ps {
		p.mu.Lock()
		p.c.SetWriteDeadline(time.Now().Add(writeWait))
		if err := p.c.WriteMessage(websocket.TextMessage, b); err != nil {
			p.mu.Unlock()
			h.mu.Lock()
			_ = p.c.Close()
			delete(h.clients, p.c)
			for subj := range h.subjects {
				delete(h.subjects[subj], p.c)
			}
			h.mu.Unlock()
			continue
		}
		p.mu.Unlock()
	}
}

func (h *Hub) Subscribe(conn *websocket.Conn, subjects []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, subj := range subjects {
		if h.subjects[subj] == nil {
			h.subjects[subj] = make(map[*websocket.Conn]*sync.Mutex)
		}
		if wmu, ok := h.clients[conn]; ok {
			h.subjects[subj][conn] = wmu
		}
	}
}

func (h *Hub) Unsubscribe(conn *websocket.Conn, subjects []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, subj := range subjects {
		if subs, ok := h.subjects[subj]; ok {
			delete(subs, conn)
			if len(subs) == 0 {
				delete(h.subjects, subj)
			}
		}
	}
}

func (h *Hub) UnsubscribeAll(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for subj := range h.subjects {
		delete(h.subjects[subj], conn)
		if len(h.subjects[subj]) == 0 {
			delete(h.subjects, subj)
		}
	}
}
