package persistence

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"moonmap.io/go-commons/helpers"
)

type ProjectDoc struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string        `bson:"name" json:"name"`
	Symbol          string        `bson:"symbol" json:"symbol"`
	Chain           string        `bson:"chain" json:"chain"`
	ContractAddress *string       `bson:"contractAddress,omitempty" json:"contractAddress,omitempty"`
	Narrative       *string       `bson:"narrative,omitempty" json:"narrative,omitempty"`
	LaunchDate      *time.Time    `bson:"launchDate,omitempty" json:"launchDate,omitempty"`
	Twitter         *string       `bson:"twitter,omitempty" json:"twitter,omitempty"`
	Telegram        *string       `bson:"telegram,omitempty" json:"telegram,omitempty"`
	Discord         *string       `bson:"discord,omitempty" json:"discord,omitempty"`
	Website         *string       `bson:"website,omitempty" json:"website,omitempty"`
	ImageUrl        string        `bson:"imageUrl" json:"imageUrl"`
	DevWallet       *string       `bson:"devWallet,omitempty" json:"devWallet,omitempty"`
	IsVerified      bool          `bson:"isVerified" json:"isVerified"`
	CreatedAt       time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time     `bson:"updatedAt" json:"updatedAt"`
	PositiveVotes   int           `bson:"positiveVotes" json:"positiveVotes"`
	NegativeVotes   int           `bson:"negativeVotes" json:"negativeVotes"`
}

func (doc *ProjectDoc) CreateNotifySubject(status string) string {
	// subject := "notify.media." + helpers.SubjPathSuffixFromKey(doc.Key)
	subject := fmt.Sprintf("notify.scope.project.%v", status)
	return subject
}

func (doc *ProjectDoc) CreateMessageId() string {
	// subject := "notify.media." + helpers.SubjPathSuffixFromKey(doc.Key)
	// subject := "notify.scope.project"
	// return subject
	msgId := doc.ID
	return msgId.Hex()
}

type MediaVariant struct {
	Key   string `bson:"key" json:"key"`
	W     int    `bson:"w" json:"w"`
	H     int    `bson:"h" json:"h"`
	Bytes int64  `bson:"bytes" json:"bytes"`
}

type MediaDoc struct {
	Key        string             `bson:"key"`
	Namespace  string             `bson:"namespace"`
	EntityID   string             `bson:"entityId"`
	Mime       string             `bson:"mime"`
	Bytes      int64              `bson:"bytes,omitempty"`
	Checksum   string             `bson:"checksum,omitempty"`
	Status     string             `bson:"status"`
	Variants   []MediaVariant     `bson:"variants"`
	Urls       map[string]string  `bson:"urls,omitempty"`
	Planned    []PlannedVariantDB `bson:"planned,omitempty"`
	UploaderID string             `bson:"uploaderId"`
	ScopeType  string             `bson:"scopeType,omitempty"`
	ScopeID    string             `bson:"scopeId,omitempty"`
	Profile    string             `bson:"profile,omitempty"`
	MediaType  string             `bson:"mediaType,omitempty"`
	ETag       string             `bson:"etag,omitempty"`

	PipelineVersion int        `bson:"pipelineVersion,omitempty"`
	Attempts        int        `bson:"attempts,omitempty"`
	NextCheckAt     *time.Time `bson:"nextCheckAt,omitempty"`
	CreatedAt       time.Time  `bson:"createdAt"`
	UpdatedAt       time.Time  `bson:"updatedAt"`
	ExpireAt        *time.Time `bson:"expireAt,omitempty"`

	Width  int `bson:"width,omitempty"`
	Height int `bson:"height,omitempty"`
}

// Exporta el plan a una forma apta para guardar en Mongo (bson-friendly)
type PlannedVariantDB struct {
	Kind string `bson:"kind" json:"kind"`
	Key  string `bson:"key" json:"key"`
	Mime string `bson:"mime" json:"mime"`
	W    int    `bson:"w" json:"w"`
	H    int    `bson:"h" json:"h"`
	Q    int    `bson:"q" json:"q"`
}

type MediaDocPartial struct {
	Key        string             `bson:"key"`
	Mime       string             `bson:"mime"`
	MediaType  string             `bson:"mediaType"` // "IMAGE|GIF|VIDEO" (si lo guardas)
	Profile    string             `bson:"profile"`
	Urls       map[string]string  `bson:"urls"`
	Planned    []PlannedVariantDB `bson:"planned"`
	UploaderID string             `bson:"uploaderId"`
}

func (doc *MediaDoc) CreateMessageId() string {
	// msgId := stream + ":" + strings.ReplaceAll(doc.Key, "/", ".") + ":" + doc.ETag
	msgId := strings.ReplaceAll(doc.Key, "/", ".") + ":" + doc.ETag
	return msgId

}

func (doc *MediaDoc) CreateNotifySubject() string {
	// subject := "notify.media." + helpers.SubjPathSuffixFromKey(doc.Key)
	subject := fmt.Sprintf("notify.media.%v.%v", doc.Namespace, doc.UploaderID)

	return subject
}

// MEDIA
// EXAMPLE: media.uploaded.image.project.projectId.avatar.projectId OR
// EXAMPLE: media.uploaded.{image|gif|video}.{user|community|project}.{userId|communityId|projectId}.{avatar|post_image}.{entityId}
func (doc *MediaDoc) CreateMediaSubject(status string) string {
	return fmt.Sprintf("media.%v.", status) +
		helpers.SubjToken(doc.MediaType) + "." + // "image" | "gif" | "video" (guárdalo en minúsculas al insertar)
		helpers.SubjToken(doc.ScopeType) + "." + // "user" | "community" | "project"
		helpers.SubjToken(doc.ScopeID) + "." +
		helpers.SubjToken(doc.Profile) + "." + // "avatar" | "post_image" | ...
		helpers.SubjToken(doc.EntityID)
}
