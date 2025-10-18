package service

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

func (s *Service) StartMetricsLogger() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				warnThresholdPct := helpers.GetEnvInt("QUEUE_WARN_THRESHOLD", 80)

				// eventos procesados en los últimos 10s
				accountCount := atomic.SwapUint64(&s.accountEvents, 0)
				logCount := atomic.SwapUint64(&s.mintEvents, 0)

				// estado actual de las colas
				progQueueLen := len(s.programSocket.Messages)
				progCapacity := cap(s.programSocket.Messages)
				logQueueLen := len(s.logSocket.Messages)
				logCapacity := cap(s.logSocket.Messages)

				logrus.Infof(
					"Processed %d account events | %d log events in last 10s | programQueue=%d/%d | logQueue=%d/%d",
					accountCount, logCount,
					progQueueLen, progCapacity,
					logQueueLen, logCapacity,
				)

				if progQueueLen*100/progCapacity > warnThresholdPct {
					message := fmt.Sprintf("⚠️ programQueue is above %d%%: %d/%d", warnThresholdPct, progQueueLen, progCapacity)
					logrus.Warn(message)
					s.AlertClient.EnqueueWarn(message)
				}
				if logQueueLen*100/logCapacity > warnThresholdPct {
					message := fmt.Sprintf("⚠️ logQueue is above %d%%: %d/%d", warnThresholdPct, logQueueLen, logCapacity)
					logrus.Warn(message)
					s.AlertClient.EnqueueWarn(message)
				}
			}
		}
	}()
}
