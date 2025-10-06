package service

import (
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
)

var rx = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\binstruction:\s*initializemint2?\b`),
	regexp.MustCompile(`(?i)\b(ix:\s*)?create\s+metadata\s+accounts?\b(\s*v[23])?`),
	regexp.MustCompile(`(?i)\bcreate\s+token\s+metadata\b`),
	regexp.MustCompile(`(?i)\b(ix:\s*)?update\s+metadata\s+accounts?\b(\s*v[23])?`),
	// Captura la invocación del programa de metadata aunque no haya “create/update” en logs
	regexp.MustCompile(`(?i)\bprogram\s+` + strings.ToLower(constants.MetadataProgramID) + `\b`),
}

func shouldProcessLogs(slot uint64, sig string, logs []string) bool {
	for _, l := range logs {
		low := strings.ToLower(l)
		for _, r := range rx {
			if r.MatchString(low) {
				logrus.WithField("signature", sig).WithField("match", r.String()).Debugf("Found log match")
				return true
			}
		}
	}

	// logrus.WithField("slot", slot).WithField("sig", sig).Info("dropping logs")
	return false
}

func (s *Service) markSeen(sig string) bool {
	if s.seen.Contains(sig) {
		return false
	}
	s.seen.Add(sig, struct{}{})
	return true
}

type ProcessorCallback func(data []byte)

func (s *Service) wsProcessor(msgs <-chan []byte, callback ProcessorCallback) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case data, ok := <-msgs:
			if !ok {
				return
			}
			s.wg.Add(1)
			go func(d []byte) {
				defer s.wg.Done()
				callback(d)
			}(data)
		}
	}
}

func (s *Service) SetStatus(st string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = st
}

func (s *Service) GetStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Service) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status == "healthy"
}
