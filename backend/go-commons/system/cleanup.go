package system

func (s *System) AddCleanUpHook(h func()) {
	s.mu.Lock()
	s.cleanUpHooks = append(s.cleanUpHooks, h)
	s.mu.Unlock()
}

func (s *System) runCleanUpHooks() {
	s.mu.Lock()
	for i := len(s.cleanUpHooks) - 1; i >= 0; i-- {
		s.cleanUpHooks[i]()
	}
	s.mu.Unlock()
}
