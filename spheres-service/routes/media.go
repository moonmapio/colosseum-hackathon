package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/system"
	"moonmap.io/spheres-service/models"
)

func MakeSphereMediaKey(sphereId, userId, ext string) string {
	return "spheres/" + sphereId + "/" + userId + "/" + ksuid.New().String() + ext
}

// POST /spheres/{sphereId}/media/presign
func PresignSphereMedia(collection *mongo.Collection, s3Config *system.S3Config, S3Presigner *s3.PresignClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}

		idHex := strings.TrimPrefix(r.URL.Path, "/spheres/")
		idHex = strings.TrimSuffix(idHex, "/media/presign")
		sid := idHex

		var req models.PresignBatchReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_REQUEST", "bad json")
			return
		}
		if len(req.Items) == 0 {
			ownhttp.WriteJSONError(w, 400, "EMPTY", "items")
			return
		}

		userId, err := bson.ObjectIDFromHex(req.UserID)
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid userId")
			return
		}

		now := time.Now()
		exp := now.Add(15 * time.Minute)

		var res models.PresignBatchRes
		var docs []interface{}

		for _, it := range req.Items {
			ext := helpers.SafeExtFromMime(it.Mime, it.Ext)
			key := MakeSphereMediaKey(sid, userId.Hex(), "."+ext)

			in := &s3.PutObjectInput{
				Bucket:      aws.String(s3Config.S3Bucket),
				Key:         aws.String(key),
				ContentType: aws.String(it.Mime),
				// ACL:         types.ObjectCannedACLPublicRead,
				// CacheControl: aws.String("public, max-age=31536000, immutable"),
			}
			ps, err := S3Presigner.PresignPutObject(r.Context(), in, func(po *s3.PresignOptions) {
				po.Expires = 15 * time.Minute
			})
			if err != nil {
				ownhttp.WriteJSONError(w, 500, "PRESIGN_FAIL", err.Error())
				return
			}

			doc := models.UploadPendingDoc{
				SphereID:  sid,
				UserID:    userId,
				Key:       key,
				Mime:      it.Mime,
				Status:    "pending",
				ExpiresAt: exp,
				CreatedAt: now,
				UpdatedAt: now,
			}
			docs = append(docs, doc)

			_, uploadURL := helpers.RewriteForClient(ps)

			publicURL := "https://s3.moonmap.io" + "/" + key

			res.Items = append(res.Items, models.PresignItemRes{
				ID:        "", // fill out after insert
				Key:       key,
				UploadURL: uploadURL,
				PublicURL: publicURL,
				ExpiresAt: exp,
			})
		}

		insert, err := collection.InsertMany(r.Context(), docs)
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "UPSERT_FAIL", err.Error())
			return
		}

		for i, id := range insert.InsertedIDs {
			oid := id.(bson.ObjectID)
			res.Items[i].ID = oid.Hex()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}
}

func CompleteSphereMedia(coll *mongo.Collection, s3Config *system.S3Config, s3client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method not allowed")
			return
		}

		var req struct {
			MediaID string `json:"mediaId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_REQUEST", "invalid json")
			return
		}
		if req.MediaID == "" {
			ownhttp.WriteJSONError(w, 400, "BAD_REQUEST", "mediaId required")
			return
		}

		oid, err := bson.ObjectIDFromHex(req.MediaID)
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid objectid for mediaId")
			return
		}

		// 1. serach pending doc
		var doc struct {
			ID        bson.ObjectID `bson:"_id"`
			Key       string        `bson:"key"`
			PublicUrl string        `bson:"publicUrl"`
			Status    string        `bson:"status"`
		}

		err = coll.FindOne(r.Context(), bson.M{"_id": oid, "status": "pending"}).Decode(&doc)
		if err != nil {
			ownhttp.WriteJSONError(w, 404, "NOT_FOUND", "pending media not found")
			return
		}

		// 2. validate on S3 object existence
		_, err = s3client.HeadObject(r.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(s3Config.S3Bucket),
			Key:    aws.String(doc.Key),
		})
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "NOT_UPLOADED", "file not uploaded yet")
			return
		}

		// 3. (Opcional) set ACL public + headers
		_, err = s3client.PutObjectAcl(r.Context(), &s3.PutObjectAclInput{
			Bucket: aws.String(s3Config.S3Bucket),
			Key:    aws.String(doc.Key),
			ACL:    types.ObjectCannedACLPublicRead,
		})
		if err != nil {
			logrus.Warnf("failed to set ACL public for %s: %v", doc.Key, err)
		}

		// cache-control copied meta data:
		// _, err = s3client.CopyObject(r.Context(), &s3.CopyObjectInput{
		//     Bucket:            aws.String(bucketName),
		//     CopySource:        aws.String(bucketName + "/" + doc.Key),
		//     Key:               aws.String(doc.Key),
		//     MetadataDirective: types.MetadataDirectiveReplace,
		//     CacheControl:      aws.String("public, max-age=31536000, immutable"),
		//     ACL:               types.ObjectCannedACLPublicRead,
		// })

		// 4. update doc on db
		_, err = coll.UpdateByID(r.Context(), doc.ID, bson.M{
			"$set": bson.M{
				"status":    "uploaded",
				"updatedAt": time.Now(),
			},
			"$unset": bson.M{"pending": ""},
		})
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "DB_ERROR", err.Error())
			return
		}

		publicUrl := helpers.GetEnv("S3_ENDPOINT_REWRITE", "") + "/" + doc.Key

		// 5. respond to the client
		ownhttp.WriteJSON(w, 200, map[string]any{
			"mediaId":   doc.ID,
			"publicUrl": publicUrl,
			"status":    "uploaded",
		})

	}
}

// DELETE /spheres/{sphereId}/media/{mediaId}
func DeleteSphereMedia(coll *mongo.Collection, s3Config *system.S3Config, s3client *s3.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodDelete {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method not allowed")
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/spheres/"), "/")
		if len(parts) < 3 || parts[1] != "media" {
			ownhttp.WriteJSONError(w, 400, "BAD_PATH", "expected /spheres/{sphereId}/media/{mediaId}")
			return
		}
		sphereId := parts[0]
		mediaId := parts[2]

		oid, err := bson.ObjectIDFromHex(mediaId)
		if err != nil {
			ownhttp.WriteJSONError(w, 400, "BAD_OBJECT_ID", "invalid mediaId")
			return
		}

		// search document
		var doc struct {
			ID  bson.ObjectID `bson:"_id"`
			Key string        `bson:"key"`
		}
		err = coll.FindOne(r.Context(), bson.M{"_id": oid, "sphereId": sphereId}).Decode(&doc)
		if err != nil {
			ownhttp.WriteJSONError(w, 404, "NOT_FOUND", "media not found")
			return
		}

		// remove from S3
		_, err = s3client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(s3Config.S3Bucket),
			Key:    aws.String(doc.Key),
		})
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "S3_DELETE_FAIL", err.Error())
			return
		}

		// remove form DB
		_, err = coll.DeleteOne(r.Context(), bson.M{"_id": oid})
		if err != nil {
			ownhttp.WriteJSONError(w, 500, "DB_DELETE_FAIL", err.Error())
			return
		}

		ownhttp.WriteJSON(w, 200, map[string]any{
			"mediaId":  mediaId,
			"sphereId": sphereId,
			"status":   "deleted",
		})

	}
}
