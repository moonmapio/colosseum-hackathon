package service

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/ownhttp"
)

// POST /waves/join
func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	ownhttp.LogRequest(r)
	if ownhttp.IsOptionsMethod(r, w) {
		return
	}

	var req struct {
		UserId   string `json:"userId"`
		SphereId string `json:"sphereId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ownhttp.WriteJSONError(w, 400, "INVALID", "bad payload")
		return
	}

	roomName := "wave:" + strings.ToLower(req.SphereId)
	at := auth.NewAccessToken(s.livekitApiKey, s.livekitApiSecret).SetIdentity(req.UserId)
	grant := &auth.VideoGrant{RoomJoin: true, Room: roomName}
	at.SetVideoGrant(grant)

	token, err := at.ToJWT()
	if err != nil {
		ownhttp.WriteJSONError(w, 500, "TOKEN_ERROR", err.Error())
		return
	}

	resp := map[string]any{
		"token":     token,
		"serverUrl": s.livekitURL,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /waves/webhook
func (s *Service) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var evt map[string]any
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		ownhttp.WriteJSONError(w, 400, "INVALID", "bad event payload")
		return
	}
	s.store.SaveEvent(r.Context(), evt)

	evType, _ := evt["event"].(string)
	room, _ := evt["room"].(string)
	sphereId := strings.TrimPrefix(room, "wave:")
	evtTime := time.Now().UTC()

	participant, _ := evt["participant"].(map[string]any)
	userId, _ := participant["identity"].(string)

	switch evType {
	case "participant_joined":
		s.store.AddParticipant(r.Context(), sphereId, userId, evtTime)
	case "participant_left":
		s.store.RemoveParticipant(r.Context(), sphereId, userId, evtTime)
	case "room_finished":
		s.store.CloseRoom(r.Context(), sphereId, evtTime)
	}

	logrus.WithFields(logrus.Fields{
		"event":  evType,
		"sphere": sphereId,
		"userId": userId,
	}).Info("webhook processed")

	w.WriteHeader(200)
}

// GET /waves/stats?sphereId=XYZ
func (s *Service) handleStats(w http.ResponseWriter, r *http.Request) {
	sphereId := r.URL.Query().Get("sphereId")
	if sphereId == "" {
		ownhttp.WriteJSONError(w, 400, "MISSING", "sphereId required")
		return
	}

	participants, err := s.GetParticipants(r.Context(), sphereId)
	if err != nil {
		ownhttp.WriteJSONError(w, 500, "LIVEKIT_ERROR", err.Error())
		return
	}

	resp := map[string]any{
		"sphereId":     sphereId,
		"active":       len(participants) > 0,
		"count":        len(participants),
		"participants": participants,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
