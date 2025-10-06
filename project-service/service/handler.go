package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/persistence"
	typesense "moonmap.io/go-commons/typesense"
)

type CreateProjectReq struct {
	ID         *string `json:"id,omitempty"`
	Name       string  `json:"name" validate:"required"`
	Symbol     string  `json:"symbol" validate:"required"`
	Chain      string  `json:"chain" validate:"required,oneof=solana"`   // validate:"required,oneof=solana ethereum bitcoin"
	LaunchDate *string `json:"launchDate,omitempty" validate:"required"` // ISO string

	ContractAddress *string `json:"contractAddress,omitempty"`
	Narrative       *string `json:"narrative,omitempty"`
	Twitter         *string `json:"twitter,omitempty" validate:"omitempty,url"`
	Telegram        *string `json:"telegram,omitempty" validate:"omitempty,url"`
	Discord         *string `json:"discord,omitempty" validate:"omitempty,url"`
	Website         *string `json:"website,omitempty" validate:"omitempty,url"`

	ImageUrl  *string `json:"imageUrl,omitempty" validate:"required,url"`
	DevWallet *string `json:"devWallet,omitempty" validate:"required,min=1"`
}

func (r *CreateProjectReq) PreNormalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Symbol = strings.ToUpper(strings.TrimSpace(r.Symbol))
	r.Chain = strings.ToLower(strings.TrimSpace(r.Chain))
}

func (r *CreateProjectReq) PostNormalize() {
	helpers.EnsureNotNilPtr(&r.ContractAddress)
	helpers.EnsureNotNilPtr(&r.Narrative)
	helpers.EnsureNotNilPtr(&r.Twitter)
	helpers.EnsureNotNilPtr(&r.Telegram)
	helpers.EnsureNotNilPtr(&r.Discord)
	helpers.EnsureNotNilPtr(&r.Website)
}

func (r *CreateProjectReq) isValid() (bool, string, error) {
	v := validator.New()
	r.PreNormalize()
	if err := v.Struct(r); err != nil {
		return false, "VALIDATION_FAILED", fmt.Errorf("review the payload you sent")
	}

	r.PostNormalize()
	if r.Name == "" || r.Symbol == "" || r.Chain == "" {
		return false, "EMPTY_NAME_SYMBOL_OR_CHAIN", fmt.Errorf("name, symbol, chain are required")
	}

	return true, "", nil
}

func (s *Service) HandleCreateOrUpdateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ownhttp.WriteJSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var in CreateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		ownhttp.WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON_REQ_DTO", "invalid json")
		return
	}

	_, code, err := in.isValid()
	if err != nil {
		ownhttp.WriteJSONError(w, http.StatusBadRequest, code, err.Error())
		return
	}

	now := time.Now().UTC()
	doc := persistence.ProjectDoc{
		Name:            in.Name,
		Symbol:          in.Symbol,
		Chain:           in.Chain,
		ContractAddress: in.ContractAddress,
		Narrative:       in.Narrative,
		Twitter:         in.Twitter,
		Telegram:        in.Telegram,
		Discord:         in.Discord,
		Website:         in.Website,
		ImageUrl:        helpers.CoalesceStr(in.ImageUrl, "https://github.com/shadcn.png"),
		DevWallet:       in.DevWallet,
		UpdatedAt:       now,
	}

	if in.LaunchDate != nil && *in.LaunchDate != "" {
		if t, err := time.Parse(time.RFC3339, *in.LaunchDate); err == nil {
			doc.LaunchDate = &t
		}
	}

	var mongoID bson.ObjectID
	if in.ID != nil {
		oid, err := bson.ObjectIDFromHex(*in.ID)
		if err == nil {
			mongoID = oid
		}
	}

	if mongoID.IsZero() {
		// === Create ===
		doc.CreatedAt = now
		res, err := s.coll.InsertOne(r.Context(), doc)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				ownhttp.WriteJSONError(w, http.StatusConflict, "DUPLICATE_KEY", "duplicate key")
				return
			}
			ownhttp.WriteJSONError(w, http.StatusBadGateway, "MONGO_ERROR", err.Error())
			return
		}
		id, err := bson.ObjectIDFromHex(helpers.IdToString(res.InsertedID))
		if err == nil {
			doc.ID = id
			go typesense.IndexProject(constants.ProjectsCollectionName, doc)
			go s.notify(&doc, "created")
		} else {
			logrus.WithError(err).Errorf("unable to create if from res.InsertedID")
		}
	} else {
		// === Update (upsert=false) ===
		filter := bson.M{"_id": mongoID}
		update := bson.M{"$set": doc}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

		res := s.coll.FindOneAndUpdate(r.Context(), filter, update, opts)
		if res.Err() != nil {
			if res.Err() == mongo.ErrNoDocuments {
				ownhttp.WriteJSONError(w, http.StatusNotFound, "NOT_FOUND", "project not found")
				return
			}
			ownhttp.WriteJSONError(w, http.StatusBadGateway, "MONGO_ERROR", res.Err().Error())
			return
		}

		if err := res.Decode(&doc); err != nil {
			ownhttp.WriteJSONError(w, http.StatusBadGateway, "MONGO_ERROR", err.Error())
			return
		}
		go typesense.IndexProject(constants.ProjectsCollectionName, doc)
		go s.notify(&doc, "updated")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(doc)
}

func (s *Service) HandleRemoveProject(w http.ResponseWriter, r *http.Request) {

	rawProjectId := r.URL.Query().Get("projectId")
	ok, projectId := ownhttp.ParseProjectId(w, rawProjectId)
	if !ok {
		return
	}

	mainLog := logrus.WithField("entityId", projectId.Hex())

	var doc persistence.ProjectDoc
	err := s.coll.FindOneAndDelete(r.Context(), bson.M{
		"_id": projectId,
	}).Decode(&doc)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			mainLog.Warnf("project %v not found", rawProjectId)
		} else {
			mainLog.Warnf("error while deleting from projects collection. %v", err.Error())
		}
	}

	s.removeMediaForProject(r.Context(), projectId)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bson.M{"status": "removed", "projectId": projectId.Hex()})
}

func (s *Service) HandleRemoveMediaForProject(w http.ResponseWriter, r *http.Request, rawProjectId string) {
	ok, projectId := ownhttp.ParseProjectId(w, rawProjectId)
	if !ok {
		return
	}

	s.removeMediaForProject(r.Context(), projectId)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bson.M{"status": "removed", "projectId": projectId.Hex()})
}

func (s *Service) removeMediaForProject(ctx context.Context, projectId *bson.ObjectID) {
	mainLog := logrus.WithField("entityId", projectId.Hex())

	// Construir keys de S3 (ejemplo: todos los prefijos bajo projects/{projectId}/)
	prefix := fmt.Sprintf("projects/%s/", projectId.Hex())

	// Listar objetos bajo ese prefijo
	listOut, err := s.S3c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.S3Cfg.S3Bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		mainLog.Warnf("error while listing prefix %v from s3 bucket %v. %v", prefix, s.S3Cfg.S3Bucket, err.Error())
	}

	if len(listOut.Contents) > 0 {
		// Preparar batch de deletes
		var objects []types.ObjectIdentifier
		for _, o := range listOut.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: o.Key})
		}

		_, err = s.S3c.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.S3Cfg.S3Bucket),
			Delete: &types.Delete{Objects: objects},
		})
		if err != nil {
			mainLog.Warnf("error while deleting objects wit key %v from s3 bucket %v. %v", prefix, s.S3Cfg.S3Bucket, err.Error())
		} else {
			mainLog.WithField("prefix", prefix).Infof("objects under prefix removed from bucket %v", s.S3Cfg.S3Bucket)
		}
	}

	res, err := s.mediaColl.DeleteMany(ctx, bson.M{
		"namespace": "projects",
		"entityId":  projectId.Hex(),
	})

	if err != nil {
		mainLog.Warnf("media assets for projectId %v not deleted. %v", projectId.Hex(), err.Error())
	} else {
		mainLog.Infof("media assets for project %v removed. deletedCount=%v", projectId.Hex(), res.DeletedCount)
		if res.DeletedCount == 0 {
			mainLog.Warnf("deleted count was 0. Requested references under %v collection were not remove. Probably the were remove before. For whatever reason they are not present anymore and that was the goal", constants.MediaAssetsCollectionName)
		}
	}

}
