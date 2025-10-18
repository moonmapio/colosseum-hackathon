package system

import "sync"

type StoreKey[T any] struct{ name string }

func NewStoreKey[T any](name string) StoreKey[T] { return StoreKey[T]{name: name} }

type storeCell struct {
	mu sync.RWMutex
	v  any
}

type Store struct {
	mu   sync.RWMutex
	data map[string]*storeCell
}

func NewStore() *Store { return &Store{data: make(map[string]*storeCell)} }

func (s *Store) cell(name string) *storeCell {
	s.mu.RLock()
	c := s.data[name]
	s.mu.RUnlock()
	if c != nil {
		return c
	}
	s.mu.Lock()
	if c = s.data[name]; c == nil {
		c = &storeCell{}
		s.data[name] = c
	}
	s.mu.Unlock()
	return c
}

func (s *Store) SetAny(name string, v any) {
	c := s.cell(name)
	c.mu.Lock()
	c.v = v
	c.mu.Unlock()
}

func (s *Store) GetAny(name string) (any, bool) {
	c := s.cell(name)
	c.mu.RLock()
	v := c.v
	c.mu.RUnlock()
	return v, v != nil
}

func (s *Store) UpdateAny(name string, fn func(old any) any) {
	c := s.cell(name)
	c.mu.Lock()
	c.v = fn(c.v)
	c.mu.Unlock()
}

func (s *Store) Delete(name string) {
	s.mu.Lock()
	delete(s.data, name)
	s.mu.Unlock()
}

func (s *Store) GetRaw(name string) (any, bool) {
	s.mu.RLock()
	c := s.data[name]
	s.mu.RUnlock()
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	v := c.v
	c.mu.RUnlock()
	return v, v != nil
}

func (s *Store) ForEach(fn func(name string, v any)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, c := range s.data {
		c.mu.RLock()
		v := c.v
		c.mu.RUnlock()
		fn(k, v)
	}
}

func Set[T any](s *Store, k StoreKey[T], v T) { s.SetAny(k.name, v) }

func Get[T any](s *Store, k StoreKey[T]) (T, bool) {
	v, ok := s.GetAny(k.name)
	if !ok {
		var zero T
		return zero, false
	}
	tv, ok := v.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return tv, true
}

func Update[T any](s *Store, k StoreKey[T], fn func(old T) T) {
	s.UpdateAny(k.name, func(old any) any {
		var o T
		if old != nil {
			if tv, ok := old.(T); ok {
				o = tv
			}
		}
		return fn(o)
	})
}

func DeleteKey[T any](s *Store, k StoreKey[T]) { s.Delete(k.name) }
