package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"moonmap.io/go-commons/ownhttp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// POST /spheres
func CreateSphere(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		var req struct {
			MintID    string `json:"mintId"`
			CreatedBy string `json:"createdBy"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_JSON", "decode")
			return
		}

		sphere := bson.M{
			"_id":         req.MintID,
			"createdBy":   req.CreatedBy,
			"createdAt":   time.Now(),
			"lastUpdated": time.Now(),
		}

		inserted, err := collection.InsertOne(r.Context(), sphere)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "INSERT_FAIL", err.Error())
			return
		}

		mintId, ok := inserted.InsertedID.(string)
		if !ok {
			ownhttp.WriteJSONError(w, 500, "INSERT_FAIL", "invalid objectid creating sphere")
			return
		}

		sphere["_id"] = mintId
		ownhttp.WriteJSON(w, 201, sphere)
	}
}
