package system

import (
	"context"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func StartMonitor(ctx context.Context, interval time.Duration) {
	go func() {
		var m runtime.MemStats
		for {
			select {
			case <-ctx.Done():
				return
			default:
				runtime.ReadMemStats(&m)
				logrus.WithFields(logrus.Fields{
					"goroutines":    runtime.NumGoroutine(),
					"allocMB":       float64(m.Alloc) / 1024 / 1024,
					"sysMB":         float64(m.Sys) / 1024 / 1024,
					"heapAllocMB":   float64(m.HeapAlloc) / 1024 / 1024,
					"gcCyclesCount": m.NumGC,
				}).Info("[STATS]")

				time.Sleep(interval)
			}
		}
	}()
}
