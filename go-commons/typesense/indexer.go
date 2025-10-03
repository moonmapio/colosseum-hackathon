package typesense

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/persistence"
)

func IndexProject(tsColl string, p persistence.ProjectDoc) error {
	tsURL := helpers.GetEnv("TYPESENSE_URL", "http://localhost:8010")
	tsKey := helpers.GetEnv("TYPESENSE_APIKEY", "xzy")

	if tsURL == "" || tsKey == "" {
		return fmt.Errorf("ensure TYPESENSE_URL and TYPESENSE_APIKEY env variables")
	}

	if tsColl == "" {
		return fmt.Errorf("no collection was defined")
	}

	var launchdate = ""
	var launchdateUnix int64
	if p.LaunchDate != nil {
		launchdate = p.LaunchDate.UTC().Format(time.RFC3339)
		launchdateUnix = p.LaunchDate.Unix()
	}

	// documento “flat” para Typesense
	payload := ProjectDoc{

		// since we are creating this document
		// we dont have an id for here like
		// the coin comming from coingecko
		ID:      p.ID.Hex(),
		MongoID: p.ID.Hex(),

		Name:            p.Name,
		Symbol:          p.Symbol,
		Chain:           p.Chain,
		ContractAddress: *p.ContractAddress,
		Narrative:       *p.Narrative,
		LaunchDate:      launchdate,
		LaunchDateUnix:  launchdateUnix,
		Twitter:         helpers.StrFromPtr(p.Twitter),
		Telegram:        helpers.StrFromPtr(p.Telegram),
		Discord:         helpers.StrFromPtr(p.Discord),
		Website:         helpers.StrFromPtr(p.Website),
		ImageUrl:        p.ImageUrl,
		IsVerified:      p.IsVerified,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
		CreatedAtUnix:   p.CreatedAt.Unix(),
		UpdatedAt:       p.UpdatedAt.Format(time.RFC3339),
		UpdatedAtUnix:   p.LaunchDate.Unix(),
		PositiveVotes:   uint32(0),
		NegativeVotes:   uint32(0),
		Source:          "moonmap",
		DevWallet:       *p.DevWallet,
	}

	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", tsURL+"/collections/"+tsColl+"/documents", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", tsKey)

	cli := &http.Client{Timeout: 1 * time.Minute}
	res, err := cli.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("typesense index failed")
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		logrus.WithField("status", res.StatusCode).Warn("typesense non-2xx")
	}

	logrus.WithFields(logrus.Fields{"name": p.Name, "mongoId": p.ID}).Info("project added to typesense successfully")

	return nil
}
