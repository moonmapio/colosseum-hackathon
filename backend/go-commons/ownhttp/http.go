package ownhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"moonmap.io/go-commons/constants"
)

type ServerOpts struct {
	ReadHeader *time.Duration
	Read       *time.Duration
	Write      *time.Duration
	Idle       *time.Duration
}

func Dur(d time.Duration) *time.Duration { return &d }

func (o *ServerOpts) withDefaults() ServerOpts {
	defRH := 3 * time.Second
	defR := 5 * time.Second
	defW := 5 * time.Second
	defI := 60 * time.Second

	out := ServerOpts{
		ReadHeader: &defRH,
		Read:       &defR,
		Write:      &defW,
		Idle:       &defI,
	}
	if o == nil {
		return out
	}
	if o.ReadHeader != nil {
		out.ReadHeader = o.ReadHeader
	}
	if o.Read != nil {
		out.Read = o.Read
	}
	if o.Write != nil {
		out.Write = o.Write
	}
	if o.Idle != nil {
		out.Idle = o.Idle
	}
	return out
}

func NewServer(ctx context.Context, srvName, bind string, handler *http.ServeMux, opts *ServerOpts) {
	o := opts.withDefaults()
	srv := &http.Server{
		Addr:              bind,
		Handler:           WithCORS(handler),
		ReadHeaderTimeout: *o.ReadHeader,
		ReadTimeout:       *o.Read,
		WriteTimeout:      *o.Write,
		IdleTimeout:       *o.Idle,
	}

	ln, err := net.Listen("tcp", bind)
	if err != nil {
		logrus.WithError(err).Fatal("cannot bind HTTP listener")
		return
	}

	if tcp, ok := ln.Addr().(*net.TCPAddr); ok {
		logrus.WithFields(logrus.Fields{
			"addr": ln.Addr().String(), // p.ej. "[::]:8080"
			"port": tcp.Port,           // p.ej. 8080 (o el asignado si era :0)
		}).Infof("%v HTTP server listening", srvName)
	} else {
		logrus.WithField("addr", ln.Addr().String()).Infof("%v HTTP server listening", srvName)
	}

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Error("http server error")
		}
	}()

	<-ctx.Done()
	_ = srv.Shutdown(ctx)
	logrus.Infof("%v shutting down server instace...", srvName)
}

func LogRequest(r *http.Request) {
	fields := logrus.Fields{
		"client_ip":   clientIP(r),
		"remote_addr": r.RemoteAddr,
		"proto":       r.Proto,
		"uri":         r.RequestURI,
	}

	// dinámicamente convertir headers → fields
	for _, h := range constants.HeaderList {
		key := strings.ReplaceAll(strings.ToLower(h), "-", "_")
		fields[key] = r.Header.Get(h)
	}

	logrus.WithFields(fields).Info("new request")
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get(constants.HeaderCFConnectingIP); v != "" {
		return v
	}

	if v := r.Header.Get(constants.HeaderXRealIP); v != "" {
		return v
	}

	if v := r.Header.Get(constants.HeaderXForwardedFor); v != "" {
		// tomar solo el primer IP en la lista
		if i := strings.IndexByte(v, ','); i > 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}

	// fallback a RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func ParseProjectId(w http.ResponseWriter, projectId string) (bool, *bson.ObjectID) {
	if projectId == "" {
		WriteJSONError(w, http.StatusBadRequest, "MISSING_PROJECT_ID", "missing projectId")
		return false, nil
	}

	_id, err := bson.ObjectIDFromHex(projectId)
	if err != nil {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_SERVER", err.Error())
		return false, nil

	}

	return true, &_id
}
