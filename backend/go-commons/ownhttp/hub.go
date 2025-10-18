package ownhttp

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 25 * time.Second
)

type Hub struct {
	mu       sync.RWMutex
	upgrader websocket.Upgrader
	Mode     string
	Clients  map[string]*SocketConnection
	Subjects map[string]map[string]*SocketConnection
	closing  bool
}

func NewHub() *Hub {
	return &Hub{
		Mode:     "all",
		Clients:  make(map[string]*SocketConnection),
		Subjects: make(map[string]map[string]*SocketConnection),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Hub) addSubsUnsafe(s *SocketConnection, subjects []string) {
	for _, subj := range subjects {
		if h.Subjects[subj] == nil {
			h.Subjects[subj] = make(map[string]*SocketConnection)
		}
		if _, ok := h.Subjects[subj][s.ID]; !ok {
			h.Subjects[subj][s.ID] = s
			s.Subs[subj] = struct{}{}
		}
	}
}

func (h *Hub) removeSubsUnsafe(s *SocketConnection, subjects []string) {
	for _, subj := range subjects {
		if subs, ok := h.Subjects[subj]; ok {
			delete(subs, s.ID)
			if len(subs) == 0 {
				delete(h.Subjects, subj)
			}
		}
		delete(s.Subs, subj)
	}
}

func (h *Hub) removeAllSubsUnsafe(s *SocketConnection) {
	if len(s.Subs) == 0 {
		return
	}
	tmp := make([]string, 0, len(s.Subs))
	for subj := range s.Subs {
		tmp = append(tmp, subj)
	}
	h.removeSubsUnsafe(s, tmp)
}

func (h *Hub) Remove(id string) {
	h.mu.RLock()
	s := h.Clients[id]
	h.mu.RUnlock()
	if s != nil {
		s.Close()
	}
}

func (h *Hub) Add(w http.ResponseWriter, r *http.Request) {
	if h.closing {
		WriteJSONError(w, http.StatusInternalServerError, "HUB_CLOSING", "hub closing")
		return
	}

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "bad request")
		return
	}
	remoteAddr := c.UnderlyingConn().RemoteAddr()
	logrus.WithField("remote_conn", remoteAddr.String()).Info("ws upgraded")

	conn := NewSocketConnection(c)
	conn.Hub = h

	h.mu.Lock()
	h.Clients[conn.ID] = conn
	h.mu.Unlock()

	conn.Init()
	go conn.Read()
	go conn.StartPinger()
}

func (h *Hub) BroadcastJSON(eventName string, data any) {
	if h.closing {
		return
	}

	payload := map[string]any{"event": eventName, "data": data}
	b, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("broadcast json marshal failed")
		return
	}

	h.mu.RLock()
	var targets map[string]*SocketConnection
	if h.Mode == "all" {
		targets = h.Clients
	} else {
		if subs, ok := h.Subjects[eventName]; ok {
			targets = subs
		} else {
			targets = nil
		}
	}

	if len(targets) == 0 {
		h.mu.RUnlock()
		return
	}

	snapshot := make([]*SocketConnection, 0, len(targets))
	for _, s := range targets {
		snapshot = append(snapshot, s)
	}
	h.mu.RUnlock()

	for _, s := range snapshot {
		if s == nil || s.Conn == nil {
			continue
		}

		s.Mutex.Lock()
		s.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := s.Conn.WriteMessage(websocket.TextMessage, b)
		s.Mutex.Unlock()
		if err != nil {
			logrus.WithError(err).Warn("broadcast failed; closing client")
			s.Close()
		}
	}
}

func (h *Hub) subscribe(s *SocketConnection, subjects []string) {
	sanitized := helpers.SanitizeStrings(subjects)
	if len(sanitized) == 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.addSubsUnsafe(s, sanitized)
	s.MainLog().WithField("op", "subscribe").WithField("subs", sanitized).Info("client subscribed")
}

func (h *Hub) unsubscribe(s *SocketConnection, subjects []string) {
	sanitized := helpers.SanitizeStrings(subjects)
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(sanitized) == 0 {
		h.removeAllSubsUnsafe(s)
	} else {
		h.removeSubsUnsafe(s, sanitized)
	}

	op := "unsubscribe_all"
	if len(sanitized) > 0 {
		op = "unsubscribe"
	}
	s.MainLog().WithField("op", op).WithField("subs", sanitized).Info("client unsubscribed")
}

func (h *Hub) GetClientsLength() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

func (h *Hub) IsSubjectMode() bool {
	return strings.EqualFold(h.Mode, "subjects")
}

func (h *Hub) Close() {
	h.closing = true
	for _, c := range h.Clients {
		c.Close()
	}
}
