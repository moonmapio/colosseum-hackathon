package publisher

import (
	"context"

	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
)

func (p *Publisher) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		p.service.Config(ctx)
		p.createPendingConsumer()

		ownhttp.NewServer(ctx, constants.S3PublisherService, sys.Bind, p.publisherRoutes(), nil)
		<-ctx.Done()
	})
}
