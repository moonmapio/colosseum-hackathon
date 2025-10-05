package service

import (
	"context"
	"strings"

	"github.com/livekit/protocol/livekit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type EmbeddedUser struct {
	ID            bson.ObjectID `bson:"_id" json:"_id"`
	FullName      string        `bson:"fullName" json:"fullName"`
	AvatarURL     string        `bson:"avatarUrl" json:"avatarUrl"`
	WalletAddress string        `bson:"walletAddress" json:"walletAddress"`
	//
}

func (s *Service) GetParticipants(ctx context.Context, sphereId string) ([]EmbeddedUser, error) {

	req := livekit.ListParticipantsRequest{Room: "wave:" + strings.ToLower(sphereId)}
	res, err := s.livekitClient.ListParticipants(ctx, &req)
	if err != nil {
		return nil, err
	}

	// extraer userIds
	ids := make([]bson.ObjectID, 0)
	for _, p := range res.GetParticipants() {
		if oid, err := bson.ObjectIDFromHex(p.Identity); err == nil {
			ids = append(ids, oid)
		}
	}

	// lookup en wallets
	proj := options.Find().SetProjection(bson.M{
		"_id":           1,
		"fullName":      1,
		"avatarUrl":     1,
		"walletAddress": 1,
		// todo add verified

	})
	cur, err := s.collWallets.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, proj)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []EmbeddedUser{}
	for cur.Next(ctx) {
		var u EmbeddedUser
		if err := cur.Decode(&u); err == nil {
			out = append(out, u)
		}
	}

	return out, nil
}
