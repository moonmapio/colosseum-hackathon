package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

func (s *Service) refreshOnce(ctx context.Context, symbols []string) {
	if len(symbols) == 0 {
		return
	}
	ctxT, cancel := context.WithTimeout(ctx, s.cg.Poller.Timeout)
	defer cancel()

	quotes, err := s.cg.Quotes(ctxT, symbols, s.vs)
	if err != nil {
		return
	}
	now := time.Now().UTC()
	for sym, price := range quotes {
		if price <= 0 {
			continue
		}
		q := Quote{Symbol: sym, Price: price, Source: s.source, At: now}
		s.store.Set(q)

		// save the price to the DB
		if !s.persist {
			logrus.Warnf("skipping persist. %s=%f", q.Symbol, q.Price)
			continue
		}

		if _, err := s.coll.InsertOne(s.ctx, q); err != nil {
			logrus.Errorf("error while saving new price of %v", sym)
			logrus.Errorln(err)
		} else {
			logrus.Infof("success saving new price of %v on mongoDB", sym)
		}
	}
}

func (s *Service) background(ctx context.Context) {
	// primer run inmediato
	s.refreshOnce(ctx, s.syms)

	// luego cada refresh (+ jitter opcional)
	t := time.NewTimer(s.refresh)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.refreshOnce(ctx, s.syms)
			next := s.refresh
			if s.jitter > 0 {
				next += time.Duration(rand.Int63n(int64(s.jitter)))
			}
			t.Reset(next)
		}
	}
}
