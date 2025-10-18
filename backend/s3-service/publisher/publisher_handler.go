package publisher

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/s3-service/core"
)

func (p *Publisher) publisherRoutes() *http.ServeMux {
	mux := ownhttp.Routes()
	if p.service.Mode == "publisher" {
		p.presign(mux)
		p.process(mux)
	}

	return mux
}

func (p *Publisher) presign(mux *http.ServeMux) {
	mux.HandleFunc("/media/presign", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)

		if ownhttp.IsOptionsMethod(r, w) {
			return
		}
		if r.Method != http.MethodPost {
			ownhttp.WriteJSONError(w, http.StatusMethodNotAllowed, "NOT_ALLOWED", "method")
			return
		}

		var req core.PresignReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "bad json")
			return
		}

		// create document under project collection
		if req.EntityID == "" || req.ScopeID == "" {
			col := persistence.MustGetCollection(constants.ProjectsCollectionName)
			result, err := col.InsertOne(r.Context(), persistence.ProjectDoc{DevWallet: &req.UserID, CreatedAt: time.Now()})
			if err != nil {
				ownhttp.WriteJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "error while creating ProjectDoc")
				return
			}

			req.EntityID = helpers.IdToString(result.InsertedID)
			req.ScopeID = req.EntityID
			logrus.Info("ProjectDoc created with id %v", req.EntityID)
		}

		// if !HasPermissions(req.UserID, req.ScopeType, req.ScopeID) {
		// 	ownhttp.WriteJSONError(w, http.StatusUnauthorized, "NOT_AUTHORIZED", "not authorized")
		// 	return
		// }

		// 1) normaliza ext y key para el presign
		ext := helpers.SafeExtFromMime(req.Mime, req.Ext)
		key := req.MakeKeyWithExt(ext)

		in := &s3.PutObjectInput{
			Bucket:      aws.String(p.service.S3Bucket),
			Key:         aws.String(key),
			ContentType: aws.String(req.Mime),
		}

		ps, err := p.service.Presigner.PresignPutObject(p.service.Ctx, in, func(po *s3.PresignOptions) {
			po.Expires = 15 * time.Minute
		})

		if err != nil {
			logrus.Errorln(err)
			ownhttp.WriteJSONError(w, http.StatusInternalServerError, "SERVER_INTERNAL", "presign")
			return
		}

		// 2) reescribe host para el cliente y arma el matrix (key, urls, plan)
		// tu lógica de rewrite → base "https://s3.moonmap.io/"
		matrix, _, uploadURL := rewriteForClient(ps, req)

		// 3) upsert (solo en insert) del documento completo (incluye urls/planned/mediaType/nextCheckAt)
		now := time.Now()
		doc := req.ToInsertDoc(matrix, now)
		_, err = p.service.Coll.UpdateOne(
			p.service.Ctx,
			bson.M{"key": doc.Key},
			bson.M{"$setOnInsert": doc},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			logrus.Errorln(err)
		}

		// 4) respuesta con event
		msgID := doc.CreateMessageId()
		res := core.BuildPresignRes(matrix, uploadURL, now.Add(15*time.Minute))
		data := core.MediaStateFromDocument(&doc)
		data.Status = "pending"
		data.Mime = req.Mime
		data.UpdatedAt = doc.CreatedAt
		data.Bytes = 0
		data.ETag = ""
		data.UploaderID = req.UserID

		res.PendingEvent = core.PresignPendingEvent{
			Stream:  "media",
			Subject: "media.pending",
			MsgID:   msgID,
			Data:    data,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)

		subject := doc.CreateNotifySubject()
		err = p.service.EventStore.PublishJSON(constants.StreamNotify, subject, msgID, res.PendingEvent.Data, nil)
		if err != nil {
			logrus.Error("failed: publishing. Verify connection to NATS server")
		}

	})
}

func (p *Publisher) process(mux *http.ServeMux) {
	mux.HandleFunc("/media/process", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)

		if r.Method != http.MethodPost {
			ownhttp.WriteJSONError(w, http.StatusMethodNotAllowed, "NOT_ALLOWED", "method")
			return
		}

		var req core.PresignPendingEvent
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ownhttp.WriteJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "bad json")
			return
		}

		mainLog := logrus.WithFields(logrus.Fields{"key": req.Data.Key, "subject": req.Subject, "stream": req.Stream, "msgId": req.MsgID})

		err := p.service.EventStore.PublishJSON(req.Stream, req.Subject, req.MsgID, req.Data, nil)
		if err != nil {
			mainLog.Error("failed: publishing. Verify connection to NATS server")
			ownhttp.WriteJSONError(w, http.StatusBadRequest, "BAD_REQUEST", "process couldnt be started")
			return

		}

		mainLog.Info("confirm: published")
		resp := map[string]any{"status": "ok"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(resp)

		return
	})
}

func rewriteForClient(ps *v4.PresignedHTTPRequest, req core.PresignReq) (matrix core.VariantsMatrix, clientBase, uploadURL string) {
	// 2) reescribe host para el cliente y arma el matrix (key, urls, plan)
	clientBase, uploadURL = helpers.RewriteForClient(ps)
	matrix = req.CreateVariantsMatrix(clientBase) // usa el mismo key normalizado y genera urls/plan

	return
}
