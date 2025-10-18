package service

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/messages"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/solana"
)

func (s *Service) OnStatus(name, status string) {
	if s.ctx.Err() != nil { // apagando
		return
	}

	msg := fmt.Sprintf("%s socket %s", name, status)
	if status == "DOWN" {
		s.AlertClient.EnqueueError(msg)
	} else {
		s.AlertClient.EnqueueInfo(msg)
	}
}

type JSONRPCReq struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func (s *Service) SubscribeLogs(conn *websocket.Conn) error {
	programs := []string{constants.TokenProgramID, constants.Token2022ProgramID, constants.MetadataProgramID}
	commitment := helpers.GetEnv("COMMITMENT", "confirmed")
	for i, prog := range programs {
		id := s.logIDBase + i
		params := []any{
			map[string]any{"mentions": []string{prog}},
			map[string]any{"commitment": commitment},
		}

		req := JSONRPCReq{JSONRPC: "2.0", ID: id, Method: "logsSubscribe", Params: params}
		s.logSocket.SendJSON(req)
		msg := fmt.Sprintf("Subscribed to logsSubscribe program=%s with id %v commitment=%v", prog, id, commitment)
		s.AlertClient.EnqueueInfo(msg)
		logrus.Infof("Subscribed logsSubscribe to program=%s id=%d commitment=%v", prog, id, commitment)
	}

	logrus.Info("Watching mint events")
	return nil
}

func (s *Service) handleLogsMessage(data []byte) {
	var ln solana.LogsNotification
	if json.Unmarshal(data, &ln) != nil || ln.Method != "logsNotification" {
		return
	}

	slot := ln.Params.Result.Context.Slot
	sig := ln.Params.Result.Value.Signature
	if sig == "" {
		return
	}

	if !s.markSeen(sig) {
		return
	}

	if shouldProcessLogs(slot, sig, ln.Params.Result.Value.Logs) {
		atomic.AddUint64(&s.mintEvents, 1)
		eventData := solana.MintCreateLog{Signature: sig, Slot: slot}
		rawData, _ := json.Marshal(eventData)

		ev := persistence.EventRecord{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Stream:    constants.StreamSolanaMints,
			Subject:   constants.SubjectSolanaLogsMintCreate,
			MsgID:     sig,
			Data:      rawData,
		}
		atomic.AddUint64(&s.logEventsSeen, 1)

		s.mu.RLock()
		st := s.status
		s.mu.RUnlock()

		switch st {
		case "healthy", "replay":
			_, err := s.publishWithRetry(ev)
			if err != nil {
				_ = s.logsBacklog.Write(ev)
				msg := fmt.Sprintf("Publish failed, msgID %v wrote to logs backlog", ev.MsgID)
				s.AlertClient.EnqueueWarn(msg)
			}
		case "nats_failed":
			_ = s.logsBacklog.Write(ev)
		default:
			logrus.Warnf("Unknown status=%s, writing logs event to backlog", st)
			_ = s.logsBacklog.Write(ev)
		}

	}

}

func (s *Service) SubscribeProgram(conn *websocket.Conn) error {
	programs := []string{constants.TokenProgramID, constants.Token2022ProgramID}
	commitment := helpers.GetEnv("COMMITMENT", "confirmed")
	for i, prog := range programs {
		id := s.programIDBase + i
		params := []any{
			prog,
			map[string]interface{}{
				"encoding":   "jsonParsed",
				"commitment": commitment,
			},
		}

		req := JSONRPCReq{JSONRPC: "2.0", ID: id, Method: "programSubscribe", Params: params}
		s.programSocket.SendJSON(req)
		msg := fmt.Sprintf("Subscribed to programSubscribe program=%s with id %v commitment=%v", prog, id, commitment)
		s.AlertClient.EnqueueInfo(msg)
		logrus.Infof("Subscribed programSubscribe to program=%s id=%d commitment=%v", prog, id, commitment)
	}

	logrus.Info("Watching program events")
	return nil
}

func (s *Service) handleProgramMessage(data []byte) {

	var n solana.ProgramNotification
	if json.Unmarshal(data, &n) != nil || n.Method != "programNotification" {
		return
	}

	slot := n.Params.Result.Context.Slot
	parsedType := n.Params.Result.Value.Account.Data.Parsed.Type
	if parsedType != "account" {
		return
	}

	atomic.AddUint64(&s.accountEvents, 1)
	pubkey := n.Params.Result.Value.Pubkey

	msgID := messages.BuildProgramMsgID(slot, pubkey, data)
	ev := persistence.EventRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Stream:    constants.StreamSolanaAccounts,
		Subject:   constants.SubjectSolanaAccountUpdated,
		MsgID:     msgID,
		Data:      data,
	}

	atomic.AddUint64(&s.programEventsSeen, 1)

	s.mu.RLock()
	st := s.status
	s.mu.RUnlock()

	switch st {
	case "healthy", "replay":
		_, err := s.publishWithRetry(ev)
		if err != nil {
			_ = s.programBacklog.Write(ev)
			msg := fmt.Sprintf("Publish failed, msgID %v wrote to backlog", ev.MsgID)
			s.AlertClient.EnqueueWarn(msg)
		}
	case "nats_failed":
		_ = s.programBacklog.Write(ev)
	default:
		logrus.Warnf("Unknown status=%s, writing to backlog as safe fallback", st)
		_ = s.programBacklog.Write(ev)
	}
}
