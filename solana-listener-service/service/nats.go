package service

import (
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/persistence"
)

func (s *Service) publishWithRetry(ev persistence.EventRecord) (jetstream.PubAckFuture, error) {
	var err error
	var pa jetstream.PubAckFuture
	backoff := []time.Duration{200 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 5 * time.Second}

	for i := 0; i < len(backoff); i++ {
		pa, err = s.EventStore.PublishBytes(ev.Stream, ev.Subject, ev.MsgID, ev.Data, nil)
		if err == nil {
			return pa, nil
		}
		select {
		case <-s.ctx.Done():
			return nil, err
		case <-time.After(backoff[i]):
		}
	}
	return nil, err
}

//	func (s *Service) publishWithRetry(ev persistence.EventRecord) error {
//		var err error
//		backoff := []time.Duration{200 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 5 * time.Second}
//		for i := 0; i < len(backoff); i++ {
//			err = s.EventStore.PublishBytesSync(ev.Stream, ev.Subject, ev.MsgID, ev.Data, nil)
//			if err == nil {
//				return nil
//			}
//			select {
//			case <-s.ctx.Done():
//				return err
//			case <-time.After(backoff[i]):
//			}
//		}
//		return err
//	}
func (s *Service) ReplayFromBacklogs() {
	s.SetStatus("replay")
	s.forceRotateBacklogs()

	// Tunables
	const (
		batchSize  = 2000 // <= tu WithPublishAsyncMaxPending(1500)
		ackTimeout = 5 * time.Second
	)

	publishBatch := func(batch []persistence.EventRecord) []bool {
		type item struct {
			i  int
			pa jetstream.PubAckFuture
		}
		results := make([]bool, len(batch))
		var inflight []item
		inflight = inflight[:0]

		enqueue := func(idx int, ev persistence.EventRecord) bool {
			pa, err := s.publishWithRetry(ev) // devuelve PubAckFuture
			if err != nil || pa == nil {
				return false
			}
			inflight = append(inflight, item{i: idx, pa: pa})
			return true
		}

		// Encolar el lote
		for i, ev := range batch {
			ok := enqueue(i, ev)
			if !ok {
				results[i] = false
			}
		}

		// Esperar ACK/ERR/timeout por cada futuro del lote
		for _, it := range inflight {
			select {
			case <-s.ctx.Done():
				results[it.i] = false
			case <-it.pa.Ok():
				results[it.i] = true
			case err := <-it.pa.Err():
				if err != nil {
					logrus.WithError(err).Debug("replay ack error")
				}
				results[it.i] = false
			case <-time.After(ackTimeout):
				results[it.i] = false
			}
		}
		return results
	}

	replayOne := func(b *persistence.Backlog) error {
		if b == nil {
			return nil
		}
		return b.ReplayBatched(batchSize, publishBatch)
	}

	if err := replayOne(s.logsBacklog); err != nil {
		// nos quedamos en replay para reintentar al próximo reconnect
		logrus.WithError(err).Warn("replay logs failed - stay in replay")
		s.SetStatus("replay")
		return
	}
	if err := replayOne(s.programBacklog); err != nil {
		logrus.WithError(err).Warn("replay program failed - stay in replay")
		s.SetStatus("replay")
		return
	}

	// ✅ todo limpio
	s.SetStatus("healthy")
	s.AlertClient.EnqueueInfo("Replay finished, service healthy")
}

func (s *Service) noPendingBacklogs() bool {
	lp := true
	pp := true
	if s.logsBacklog != nil {
		lp = !s.logsBacklog.HasPending()
	}
	if s.programBacklog != nil {
		pp = !s.programBacklog.HasPending()
	}
	return lp && pp
}

func (s *Service) forceRotateBacklogs() {
	if s.logsBacklog != nil {
		_ = s.logsBacklog.RotateNow()
	}
	if s.programBacklog != nil {
		_ = s.programBacklog.RotateNow()
	}
}
