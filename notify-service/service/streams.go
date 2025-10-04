package service

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"moonmap.io/go-commons/constants"
)

func (s *Service) CreateStreamMedia() {

	// ###############################################################################################################
	// ###############################################################################################################
	// ###############################################################################################################

	// media.uploaded.<mediaType>.<scopeType>.<scopeId>.<profile>.<entityId>
	// media.process.started.<mediaId>
	// media.process.completed.<mediaId>
	// media.process.failed.<mediaId>
	// media.reprocess.<mediaId>
	// media.delete.<mediaId>

	// mediaType: image|gif|video
	// scopeType: user|community|project
	// profile: avatar|post_image|post_video|logo|banner|cove

	// this could be moved to another service who just configures nats
	// s.eventStore.CreateStreamWithSubjects("media", []string{"media.*"})
	streamName := constants.StreamMedia
	s.EventStore.CreateStreamWithSubjects(s.Ctx, streamName, []string{
		fmt.Sprintf("%s.pending", streamName),
		fmt.Sprintf("%s.pending.>", streamName),

		fmt.Sprintf("%s.uploaded", streamName),   // disparo tras HEAD OK
		fmt.Sprintf("%s.uploaded.>", streamName), // disparo tras HEAD OK

		fmt.Sprintf("%s.process", streamName),   // started/completed/failed por mediaId
		fmt.Sprintf("%s.process.>", streamName), // started/completed/failed por mediaId

		fmt.Sprintf("%s.reprocess", streamName),   // cmd para re-procesar
		fmt.Sprintf("%s.reprocess.>", streamName), // cmd para re-procesar

		fmt.Sprintf("%s.delete", streamName),   // cmd para borrar original/variantes
		fmt.Sprintf("%s.delete.>", streamName), // cmd para borrar original/variantes
	})
	s.enqueueStreamCreation(streamName)
}

func (s *Service) CreateStreamNotify() {

	// ###############################################################################################################
	// ###############################################################################################################
	// ###############################################################################################################

	// notify.user.<userId>
	// notify.scope.<scopeType>.<scopeId>
	// notify.media.<mediaId>   // publicar con Nats-Rollup: sub
	// opcional: notify.post.<postId>

	// this could be moved to another service who just configures nats
	streamName := constants.StreamNotify
	s.EventStore.CreateStreamWithSubjects(s.Ctx, streamName, []string{
		fmt.Sprintf("%s.user", streamName),    // canal por-usuario
		fmt.Sprintf("%s.user.>", streamName),  // canal por-usuario
		fmt.Sprintf("%s.scope", streamName),   // canal por-scope (community/project/user)
		fmt.Sprintf("%s.scope.>", streamName), // canal por-scope (community/project/user)

		fmt.Sprintf("%s.media", streamName),
		fmt.Sprintf("%s.media.>", streamName),

		// opcional: "notify.post.*"    // canal por-post si te interesa
	})
	s.enqueueStreamCreation(streamName)

	// eventStore = system.NewEventStore("s3-service")

	// ###############################################################################################################
	// ###############################################################################################################
	// ###############################################################################################################

	// ###############################################################################################################
	// ###############################################################################################################
	// ###############################################################################################################

	// content
	// content.community.created|updated|deleted.<communityId>
	// content.post.created|updated|deleted.<communityId>.<postId>
	// content.comment.created|deleted.<communityId>.<postId>.<commentId>
	// content.reaction.added|removed.<target>.<communityId>.<postId>.<userId> // target: post|comment|project

	// this could be moved to another service who just configures nats
	// s.EventStore.CreateStreamWithSubjects("content", []string{
	// 	"content.community.>", // created/updated/deleted.<communityId>
	// 	"content.post.>",      // created/updated/deleted.<communityId>.<postId>
	// 	"content.comment.>",   // created/deleted.<communityId>.<postId>.<commentId>
	// 	"content.reaction.>",  // added/removed.<target>.<ids...>
	// })

}

func (s *Service) CreateStreamRequest() {
	streamName := constants.StreamRequests
	s.EventStore.CreateStreamWithSubjects(s.Ctx, streamName, []string{
		fmt.Sprintf("%s.>", streamName),
	})
	s.enqueueStreamCreation(streamName)
}

func (s *Service) CreateStreamReflector() {
	streamName := constants.StreamReflector
	s.EventStore.CreateStreamWithSubjects(s.Ctx, streamName, []string{
		fmt.Sprintf("%s.>", streamName),
	})
	s.enqueueStreamCreation(streamName)
}

func (s *Service) CreateStreamSolana() {
	config := jetstream.StreamConfig{
		Name:        constants.StreamSolanaMints,
		Retention:   jetstream.LimitsPolicy,
		Storage:     jetstream.FileStorage,
		Duplicates:  1 * time.Hour,
		AllowRollup: true,
		Replicas:    1,
		MaxAge:      7 * 24 * time.Hour,
		MaxBytes:    -1, //1073741824 * 2, // 1GB
		Subjects:    []string{fmt.Sprintf("%s.>", constants.StreamSolanaMints)},
	}

	s.EventStore.CreateStreamWithConfig(s.Ctx, config)
	s.enqueueStreamCreation(config.Name)

	config = jetstream.StreamConfig{
		Name:        constants.StreamSolanaAccounts,
		Retention:   jetstream.LimitsPolicy,
		Storage:     jetstream.FileStorage,
		Duplicates:  1 * time.Hour,
		AllowRollup: true,
		Replicas:    1,
		MaxAge:      2 * time.Hour,
		MaxBytes:    -1, //1073741824 * 2, // 1GB
		Subjects:    []string{fmt.Sprintf("%s.>", constants.StreamSolanaAccounts)},
	}

	s.EventStore.CreateStreamWithConfig(s.Ctx, config)
	s.enqueueStreamCreation(config.Name)
	s.EventStore.GetConn()
}

func (s *Service) enqueueStreamCreation(streamName string) {
	msg := fmt.Sprintf("Stream %s created or updated", streamName)
	s.QueueManager.EnqueueInfo(s.QueueManager.ServiceName, msg)
}
