package core

import (
	"strings"
	"time"

	"moonmap.io/go-commons/persistence"
)

type PresignReq struct {
	Namespace string `json:"namespace"` // "projects", "communities", "users"	-> separacion principal
	ScopeID   string `json:"scopeId"`   // userId|communityId|projectId 		-> id del dueño (userId, communityId, projectId)
	Profile   string `json:"profile"`   // avatar|post_image|post_video| 		-> que uso tiene?
	EntityID  string `json:"entityId"`  // 										-> id lógico de la pieza (postId, etc.)

	ScopeType string `json:"scopeType"` // user|community|project|banner
	UserID    string `json:"userId"`    // createdBy
	Ext       string `json:"ext"`
	Mime      string `json:"mime"`
}

// Key schema
// <namespace>/<scopeId>/<profile>/<entityId>/v1/original.<ext>
// communities/community123/post_image/post789/v1/original.jpg
// communities/community123/post_video/post789/v1/original.mp4
// projects/project123/avatar/project123/v1/original.png
// users/user123/avatar/user123/v1/original.png
// users/user123/banner/user123/v1/original.png

type PresignPendingEvent struct {
	Stream  string     `json:"stream"`
	Subject string     `json:"subject"`
	MsgID   string     `json:"msgId"`
	Data    MediaState `json:"data"`
}

type PresignRes struct {
	Key          string              `json:"key"`
	UploadURL    string              `json:"uploadUrl"`
	ExpiresAt    time.Time           `json:"expiresAt"`
	MediaID      string              `json:"mediaId"`
	Urls         map[string]string   `json:"urls"`
	Status       string              `json:"status"`
	PendingEvent PresignPendingEvent `json:"pending_event"`
}

type ConfirmReq struct {
	Key    string `json:"key"`
	Bytes  int64  `json:"bytes"`
	Mime   string `json:"mime"`
	UserID string `json:"userId"`
}

func (r PresignReq) ToInsertDoc(matrix VariantsMatrix, now time.Time) persistence.MediaDoc {
	return persistence.MediaDoc{
		Key:         matrix.Key,
		Namespace:   r.Namespace,
		EntityID:    r.EntityID,
		Mime:        r.Mime,
		UploaderID:  r.UserID,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
		Attempts:    0,
		NextCheckAt: &now,

		// contexto
		ScopeType: strings.ToLower(r.ScopeType),
		ScopeID:   r.ScopeID,
		Profile:   strings.ToLower(r.Profile),

		// derivados
		MediaType:       matrix.MediaType,
		Urls:            matrix.Urls,
		Planned:         matrix.PlannedForDB(), // arreglo con {kind,key,mime,w,h,q}
		PipelineVersion: PipelineVersion,
	}
}

func BuildPresignRes(m VariantsMatrix, uploadURL string, expiresAt time.Time) PresignRes {
	return PresignRes{
		MediaID:   m.Key,
		Key:       m.Key,
		Urls:      m.Urls,
		Status:    "pending",
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
	}
}

type MediaState struct {
	Key          string    `json:"key"`
	Mime         string    `json:"mime"`
	UploaderID   string    `json:"uploaderId"`
	Status       string    `json:"status"` // pending|uploaded|processing|ready|failed
	ETag         string    `json:"etag,omitempty"`
	Bytes        int64     `json:"bytes,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt"`
	TransitionOk bool      `json:"transitionOk"`
}

func MediaStateFromDocument(doc *persistence.MediaDoc) MediaState {
	return MediaState{
		Key:          doc.Key,
		Mime:         doc.Mime,
		UploaderID:   doc.UploaderID,
		Status:       doc.Status,
		ETag:         doc.ETag,
		Bytes:        doc.Bytes,
		UpdatedAt:    doc.UpdatedAt,
		TransitionOk: true,
	}
}
