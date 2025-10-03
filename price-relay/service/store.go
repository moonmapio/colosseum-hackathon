package service

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Quote struct {
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Source string    `json:"source"`
	At     time.Time `json:"at"`
}

type Store struct {
	mu   sync.RWMutex
	data map[string]Quote
	ttl  time.Duration
}

func NewStore(ttl time.Duration) *Store {
	return &Store{data: map[string]Quote{}, ttl: ttl}
}

func (s *Store) Get(sym string) (Quote, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q, ok := s.data[sym]
	if !ok || time.Since(q.At) > s.ttl {
		return Quote{}, false
	}
	return q, true
}

func (s *Store) Set(q Quote) {
	s.mu.Lock()
	s.data[q.Symbol] = q
	s.mu.Unlock()
	logrus.Infof("Local store updated with new prices for %v comming from %v", q.Symbol, q.Source)
}
