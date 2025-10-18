package consumer

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
)

func (c *Consumer) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		c.service.Config(ctx)

		var wg sync.WaitGroup
		for i := 0; i < c.service.Workers; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for media := range c.service.MediaChannel {
					logrus.WithFields(logrus.Fields{
						"worker":     id,
						"key":        media.Key,
						"etag":       media.ETag,
						"uploaderId": media.UploaderID,
					}).Info("processing...")
					c.process(media)
				}
			}(i)
		}

		go c.createTransformConsumer()
		ownhttp.NewServer(ctx, constants.S3ConsumerServiceName, sys.Bind, c.consumerRoutes(), nil)

		<-ctx.Done()
		close(c.service.MediaChannel)
		wg.Wait()
		return
	})
}
