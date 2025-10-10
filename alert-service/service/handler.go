package service

import (
	"encoding/json"
	"net/http"

	"moonmap.io/go-commons/messages"
	"moonmap.io/go-commons/ownhttp"
)

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()

	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)
		if ownhttp.IsOptionsMethod(r, w) {
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var msg messages.EventMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			ownhttp.WriteJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
			return
		}

		s.QueueManager.EnqueueMessage(msg.ServiceName, msg.Message, msg.Level)
		ownhttp.WriteJSON(w, http.StatusAccepted, map[string]any{
			"message": "accepted",
		})
	})

	return mux
}
