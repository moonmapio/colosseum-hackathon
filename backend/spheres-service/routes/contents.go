package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
	"moonmap.io/spheres-service/models"
)

var usersLookup = bson.D{
	{Key: "from", Value: "wallets"},
	{Key: "let", Value: bson.D{{Key: "uid", Value: "$userId"}}},
	{Key: "pipeline", Value: mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{"$_id", "$$uid"}}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "fullName", Value: 1},
			{Key: "avatarUrl", Value: 1},
			// todo add verified

		}}},
	}},
	{Key: "as", Value: "user"},
}

var endProjection = bson.D{
	{Key: "_id", Value: 1},
	{Key: "sphereId", Value: 1},
	{Key: "type", Value: 1},
	{Key: "text", Value: 1},
	{Key: "mediaUrls", Value: 1},
	{Key: "reactions", Value: 1},
	// {
	// 	Key: "reactions", Value: bson.D{
	// 		{Key: "$arrayToObject", Value: bson.D{
	// 			{Key: "$map", Value: bson.D{
	// 				{Key: "input", Value: bson.D{{Key: "$objectToArray", Value: "$reactions"}}},
	// 				{Key: "as", Value: "r"},
	// 				{Key: "in", Value: bson.D{
	// 					{Key: "k", Value: "$$r.k"},
	// 					{Key: "v", Value: bson.D{{Key: "$size", Value: "$$r.v"}}},
	// 				}},
	// 			}},
	// 		}},
	// 	},
	// },
	{Key: "parentId", Value: 1},
	{Key: "createdAt", Value: 1},
	{Key: "updatedAt", Value: 1},
	{Key: "deleted", Value: 1},
	{Key: "user", Value: 1},
}

// 1. Posts raíz con preview de replies
// GET /spheres/{sphereId}/contents?limit=20&after=hexId
func GetSpherePosts(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/spheres/"), "/")
		if len(parts) < 2 || parts[1] != "contents" {
			ownhttp.WriteJSONError(w, 400, "BAD_PATH", "invalid path")
			return
		}
		sphereId := parts[0]

		limit := int64(10)
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 10 {
			limit = int64(l)
		}

		afterStr := r.URL.Query().Get("after")
		var afterID bson.ObjectID
		if afterStr != "" {
			var err error
			afterID, err = bson.ObjectIDFromHex(afterStr)
			if err != nil {
				ownhttp.WriteJSONError(w, 400, "BAD_AFTER", "invalid after id")
				return
			}
		}

		match := bson.D{
			{Key: "sphereId", Value: sphereId},
			{Key: "deleted", Value: false},
			{Key: "parentId", Value: nil},
		}
		if afterStr != "" {
			match = append(match, bson.E{
				Key:   "_id",
				Value: bson.D{{Key: "$lt", Value: afterID}},
			})
		}

		parentsPipeline := mongo.Pipeline{
			{{Key: "$match", Value: match}},
			{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
			{{Key: "$limit", Value: limit}},
			{{Key: "$lookup", Value: usersLookup}},
			{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$user"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}}},
			{{Key: "$project", Value: endProjection}},
		}

		cur, err := collection.Aggregate(r.Context(), parentsPipeline)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "QUERY_FAIL", err.Error())
			return
		}
		defer cur.Close(r.Context())

		parents := []bson.M{}
		if err := cur.All(r.Context(), &parents); err != nil {
			ownhttp.WriteJSONError(w, 500, "CURSOR_FAIL", err.Error())
			return
		}

		if len(parents) == 0 {
			ownhttp.WriteJSON(w, 200, bson.M{
				"parents":   []bson.M{},
				"childrens": bson.M{},
			})
			return
		}

		parentIDs := make([]bson.ObjectID, 0, len(parents))
		for _, p := range parents {
			if id, ok := p["_id"].(bson.ObjectID); ok {
				parentIDs = append(parentIDs, id)
			}
		}

		childrenPipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{
				{Key: "sphereId", Value: sphereId},
				{Key: "deleted", Value: false},
				{Key: "parentId", Value: bson.D{{Key: "$in", Value: parentIDs}}},
			}}},
			{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
			{{Key: "$lookup", Value: usersLookup}},
			{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$user"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}}},
			{{Key: "$project", Value: endProjection}},
			{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$parentId"},
				{Key: "children", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
			}}},
			{{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "parentId", Value: "$_id"},
				{Key: "children", Value: bson.D{{Key: "$slice", Value: bson.A{"$children", 3}}}},
			}}},
		}

		cur2, err := collection.Aggregate(r.Context(), childrenPipeline)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "QUERY_FAIL_CHILDREN", err.Error())
			return
		}
		defer cur2.Close(r.Context())

		type grouped struct {
			ParentID any      `bson:"parentId"`
			Children []bson.M `bson:"children"`
		}
		groupedRows := []grouped{}
		if err := cur2.All(r.Context(), &groupedRows); err != nil {
			ownhttp.WriteJSONError(w, 500, "CURSOR_FAIL_CHILDREN", err.Error())
			return
		}

		childrens := bson.M{}
		for _, g := range groupedRows {
			switch v := g.ParentID.(type) {
			case string:
				childrens[v] = g.Children
			case bson.ObjectID:
				childrens[v.Hex()] = g.Children
			default:
				childrens[fmt.Sprintf("%v", v)] = g.Children
			}
		}

		ownhttp.WriteJSON(w, 200, bson.M{
			"parents":   parents,
			"childrens": childrens,
		})

	}
}

// 2. Replies of a specific post
// /spheres/{sphereId}/contents/{contentId}/replies?limit=20&after=hexId
func GetSphereReplies(collection *mongo.Collection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/spheres/"), "/")
		if len(parts) < 4 || parts[1] != "contents" || parts[2] == "" || parts[3] != "replies" {
			ownhttp.WriteJSONError(w, 400, "BAD_PATH", "invalid path")
			return
		}
		sphereId := parts[0]
		parentId, err := bson.ObjectIDFromHex(parts[2]) // contentId of parent post
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_PARENT_ID", "invalid parent id")
			return
		}

		limit := int64(20)
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
			limit = int64(l)
		}

		afterStr := r.URL.Query().Get("after")
		var afterID bson.ObjectID
		if afterStr != "" {
			var err error
			afterID, err = bson.ObjectIDFromHex(afterStr)
			if err != nil {
				ownhttp.WriteJSONError(w, 400, "BAD_AFTER", "invalid after id")
				return
			}
		}

		match := bson.D{
			{Key: "sphereId", Value: sphereId},
			{Key: "deleted", Value: false},
			{Key: "parentId", Value: parentId},
		}
		if afterStr != "" {
			match = append(match, bson.E{Key: "_id", Value: bson.D{{Key: "$lt", Value: afterID}}})
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: match}},
			{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
			{{Key: "$limit", Value: limit}},
			{{Key: "$lookup", Value: usersLookup}},
			{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$user"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}}},
		}

		cur, err := collection.Aggregate(r.Context(), pipeline)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "QUERY_FAIL", err.Error())
			return
		}
		defer cur.Close(r.Context())

		docs := []bson.M{}
		if err := cur.All(r.Context(), &docs); err != nil {
			ownhttp.WriteJSONError(w, 500, "CURSOR_FAIL", err.Error())
			return
		}

		ownhttp.WriteJSON(w, 200, docs)
	}
}

//	{
//	  "id": "contentId",
//	  "sphereId": "sphereId",
//	  "userId": "u123",
//	  "parentId": null,
//	  "type": "text",
//	  "text": "mensaje",
//	  "mediaUrls": [],
//	  "createdAt": "2025-09-10T12:00:00Z",
//	  "updatedAt": "2025-09-10T12:00:00Z",
//	  "deleted": false
//	}
//
// POST /spheres/{sphereId}/contents
func CreateSphereContent(collection *mongo.Collection, mediaCollection *mongo.Collection, eventStore *system.NatsEventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		idHex := strings.TrimPrefix(r.URL.Path, "/spheres/")
		idHex = strings.TrimSuffix(idHex, "/contents")
		sid := idHex // mintId - sphereId

		var req struct {
			UserID   string   `json:"userId"`
			ParentID *string  `json:"parentId"`
			Type     string   `json:"type"`
			Text     string   `json:"text"`
			MediaIDs []string `json:"mediaIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_REQUEST", "invalid json")
			return
		}

		userId, err := bson.ObjectIDFromHex(req.UserID)
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid objectid userId")
			return
		}

		var parentId *bson.ObjectID
		if req.ParentID != nil {
			oid, err := bson.ObjectIDFromHex(*req.ParentID)
			if err != nil {
				ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid objectid for parentId")
				return
			}
			parentId = &oid
		}

		mediaOids := make([]bson.ObjectID, 0, len(req.MediaIDs))
		for _, mid := range req.MediaIDs {
			oid, err := bson.ObjectIDFromHex(mid)
			if err != nil {
				ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid objectid for mediaId")
				return
			}
			mediaOids = append(mediaOids, oid)
		}

		cur, err := mediaCollection.Find(r.Context(), bson.M{
			"_id":    bson.M{"$in": mediaOids},
			"status": "uploaded",
		})
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "DB_ERROR", err.Error())
			return
		}
		defer cur.Close(r.Context())

		type mediaDoc struct {
			ID  bson.ObjectID `bson:"_id"`
			Key string        `bson:"key"`
		}
		var found []mediaDoc
		if err := cur.All(r.Context(), &found); err != nil {
			ownhttp.WriteJSONError(w, 500, "CURSOR_FAIL", err.Error())
			return
		}

		if len(found) != len(mediaOids) {
			ownhttp.WriteJSONError(w, 400, "MEDIA_INVALID", "some media not uploaded or expired")
			return
		}

		mediaUrls := make([]string, 0, len(found))
		for _, m := range found {
			publicUrl := fmt.Sprintf("https://%s/%s", "s3.moonmap.io", m.Key) // ajusta host base
			mediaUrls = append(mediaUrls, publicUrl)
		}

		now := time.Now()
		doc := bson.M{
			"sphereId":  sid,
			"userId":    userId,
			"type":      req.Type,
			"text":      req.Text,
			"mediaUrls": mediaUrls,
			"reactions": bson.M{},
			"createdAt": now,
			"updatedAt": now,
			"deleted":   false,
			"parentId":  parentId,
		}

		inserted, err := collection.InsertOne(r.Context(), doc)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "INSERT_FAIL", err.Error())
			return
		}
		oid := inserted.InsertedID.(bson.ObjectID)

		if len(mediaOids) > 0 {
			// mark all as attached
			_, err = mediaCollection.UpdateMany(r.Context(),
				bson.M{"_id": bson.M{"$in": mediaOids}},
				bson.M{"$set": bson.M{"status": "attached", "updatedAt": time.Now()}},
			)
			if err != nil {
				logrus.Warnf("failed to update media to attached: %v", err)
			}
		}

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{
				{Key: "_id", Value: oid},
				{Key: "deleted", Value: false},
			}}},
			{{Key: "$lookup", Value: usersLookup}},
			{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$user"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}}},
			{{Key: "$project", Value: endProjection}},
		}

		cur, err = collection.Aggregate(r.Context(), pipeline)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "INSERT_FAIL", err.Error())
			return
		}
		defer cur.Close(r.Context())

		docs := []models.SphereContentCreated{}
		if err := cur.All(r.Context(), &docs); err != nil {
			ownhttp.WriteJSONError(w, 500, "CURSOR_FAILED", err.Error())
			return
		}

		if len(docs) == 0 {
			ownhttp.WriteJSONError(w, 404, "NOT_INSERTED", "not found after insert")
			return
		}

		insertedWithUser := docs[0]

		messageId := ksuid.New()
		subject := "spheres.content.added." + idHex
		_ = eventStore.PublishJSON(constants.StreamSpheres, subject, messageId.String(), insertedWithUser, nil)

		ownhttp.WriteJSON(w, 201, insertedWithUser)

	}
}

//	{
//	  "id": "contentId",
//	  "sphereId": "sphereId",
//	  "updates": { "text": "nuevo texto" },
//	  "updatedAt": "2025-09-10T12:30:00Z"
//	}
//
// PATCH /spheres/{sphereId}/contents/{contentId}
func UpdateSphereContent(collection *mongo.Collection, eventStore *system.NatsEventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/spheres/"), "/")
		if len(parts) < 3 || parts[1] != "contents" {
			ownhttp.WriteJSONError(w, 400, "BAD_PATH", "invalid path")
			return
		}
		sphereIdHex := parts[0]
		contentId, _ := bson.ObjectIDFromHex(parts[2])

		var req struct {
			Text      *string   `json:"text"`
			MediaUrls *[]string `json:"mediaUrls"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		now := time.Now()
		updates := make(map[string]interface{})
		updates["updatedAt"] = now
		if req.Text != nil {
			updates["text"] = *req.Text
		}
		if req.MediaUrls != nil {
			updates["mediaUrls"] = *req.MediaUrls
		}

		_, err := collection.UpdateByID(r.Context(), contentId, bson.M{"$set": updates})
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "UPDATE_FAIL", err.Error())
			return
		}

		evt := models.SphereContentUpdated{
			ID:        contentId.Hex(),
			SphereID:  sphereIdHex,
			Updates:   updates,
			UpdatedAt: now,
		}

		messageId := ksuid.New()
		subject := "spheres.content.updated." + sphereIdHex
		_ = eventStore.PublishJSON(constants.StreamSpheres, subject, messageId.String(), evt, nil)

		ownhttp.WriteJSON(w, 200, evt)
	}
}

//	{
//	  "id": "contentId",
//	  "sphereId": "sphereId",
//	  "deleted": true,
//	  "updatedAt": "2025-09-10T12:40:00Z"
//	}
//
// DELETE /spheres/{sphereId}/contents/{contentId}
func DeleteSphereContent(collection *mongo.Collection, eventStore *system.NatsEventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/spheres/"), "/")
		if len(parts) < 3 || parts[1] != "contents" {
			ownhttp.WriteJSONError(w, 400, "BAD_PATH", "invalid path")
			return
		}
		sphereIdHex := parts[0]
		contentId, _ := bson.ObjectIDFromHex(parts[2])

		now := time.Now()
		update := bson.M{
			"$set": bson.M{
				"deleted":   true,
				"updatedAt": now,
			},
		}
		_, err := collection.UpdateByID(r.Context(), contentId, update)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "DELETE_FAIL", err.Error())
			return
		}

		evt := models.SphereContentDeleted{
			ID:        contentId.Hex(),
			SphereID:  sphereIdHex,
			Deleted:   true,
			UpdatedAt: now,
		}

		messageId := ksuid.New()
		subject := "spheres.content.deleted." + sphereIdHex
		_ = eventStore.PublishJSON(constants.StreamSpheres, subject, messageId.String(), evt, nil)

		ownhttp.WriteJSON(w, 200, evt)
	}
}
