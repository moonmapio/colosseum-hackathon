package routes

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
	"moonmap.io/spheres-service/models"
)

// PATCH /spheres/{sphereId}/contents/{contentId}/reactions
func ReactSphereContent(collection *mongo.Collection, eventStore *system.NatsEventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch && r.Method != http.MethodPost && r.Method != http.MethodDelete {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/spheres/"), "/")
		if len(parts) < 4 || parts[1] != "contents" || parts[3] != "reactions" {
			ownhttp.WriteJSONError(w, 400, "BAD_PATH", "invalid path")
			return
		}
		sphereIdHex := parts[0]
		contentId, err := bson.ObjectIDFromHex(parts[2])
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid content id")
			return
		}

		var req struct {
			UserID   string  `json:"userId"`   // requiered for audit
			Symbol   string  `json:"symbol"`   // emoji or reactions code
			ParentID *string `json:"parentId"` // hex o nil

		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_JSON", err.Error())
			return
		}

		userIdObjectID, err := bson.ObjectIDFromHex(req.UserID)
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "OBJECT_ID", err.Error())
			return
		}

		if req.ParentID != nil {
			_, err = bson.ObjectIDFromHex(*req.ParentID)
			if err != nil {
				ownhttp.WriteJSONError(w, 400, "OBJECT_ID", err.Error())
				return
			}
		}

		// build update
		action := "added"
		var update bson.M
		if r.Method == http.MethodDelete {
			update = bson.M{"$pull": bson.M{"reactions." + req.Symbol: req.UserID}}
			action = "removed"
		} else {
			update = bson.M{"$addToSet": bson.M{"reactions." + req.Symbol: req.UserID}}
		}

		filter := bson.M{"_id": contentId}
		if req.ParentID != nil {
			parentOID, err := bson.ObjectIDFromHex(*req.ParentID)
			if err != nil {
				ownhttp.WriteJSONError(w, 400, "BAD_PARENT_ID", err.Error())
				return
			}
			filter["parentId"] = parentOID
		} else {
			filter["$or"] = []bson.M{
				{"parentId": bson.M{"$exists": false}},
				{"parentId": nil},
			}
		}

		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		result := models.SphereContentReaction{}

		err = collection.FindOneAndUpdate(r.Context(), filter, update, opts).Decode(&result)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "REACTION_FAIL", err.Error())
			return
		}

		result.Action = action
		result.Symbol = req.Symbol
		result.UserID = userIdObjectID

		messageId := ksuid.New()
		subject := "spheres.content.reacted." + sphereIdHex
		_ = eventStore.PublishJSON(constants.StreamSpheres, subject, messageId.String(), result, nil)

		ownhttp.WriteJSON(w, 200, result)
	}
}
