package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"moonmap.io/go-commons/constants"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/ownhttp"
	"moonmap.io/go-commons/persistence"
	"moonmap.io/go-commons/system"
	"moonmap.io/price-relay/providers"
)

type Service struct {
	store   *Store
	cg      *providers.ProviderClient
	syms    []string
	vs      string
	refresh time.Duration
	jitter  time.Duration
	source  string
	ctx     context.Context
	coll    *mongo.Collection
	persist bool
}

func New() *Service {
	ttl := helpers.GetEnvDur("TTL", 30*time.Second)
	s := &Service{
		store:   NewStore(ttl),
		cg:      providers.NewProviderClient(helpers.GetEnvDur("HTTP_TIMEOUT", 3*time.Second)),
		syms:    helpers.SplitCSV(helpers.GetEnv("SYMBOLS", "SOL,BTC")),
		vs:      helpers.GetEnv("VS", "USD"),
		refresh: helpers.GetEnvDur("REFRESH_EVERY", 20*time.Second),
		jitter:  helpers.GetEnvDur("JITTER", 3*time.Second),
		source:  helpers.GetEnvOrFail("SOURCE"),
	}

	s.persist = helpers.GetEnv("PERSIST", "true") == "true"
	return s
}

func (s *Service) Config(ctx context.Context) {
	s.ctx = ctx
	s.coll = persistence.MustGetCollection(constants.PriceTicksCollectionName)

	// Indexes básicos
	_, err := s.coll.Indexes().CreateMany(s.ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "symbol", Value: 1}, {Key: "vs", Value: 1}, {Key: "source", Value: 1}, {Key: "at", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "symbol", Value: 1}, {Key: "at", Value: -1}}},
	})

	if err != nil {
		logrus.Fatal(err)
	}
}

func (s *Service) Start(sys *system.System) {
	sys.Run(func(ctx context.Context) {
		s.Config(ctx)

		// refresco en background con el poller reutilizando ownhttp
		go s.background(ctx)

		// http server
		ownhttp.NewServer(ctx, "price-relay", sys.Bind, s.routes(), nil)
	})
}

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()
	mux.HandleFunc("/prices", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)
		if ownhttp.IsOptionsMethod(r, w) {
			return
		}

		s.handlePrices(w, r)
	})
	return mux
}

func (s *Service) handlePrices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("symbols")
	vs := r.URL.Query().Get("vs")
	if vs == "" {
		vs = s.vs
	}
	if !strings.EqualFold(vs, s.vs) {
		ownhttp.WriteJSONError(w, http.StatusBadRequest, "NOT_SUPPORTED", "vs not supported")
		return
	}
	syms := s.syms
	if q != "" {
		syms = helpers.SplitCSV(q)
	}

	syms = helpers.UniqueUpper(syms)
	max := helpers.GetEnvInt("MAX_SYMBOLS", 20)
	if len(syms) > max {
		syms = syms[:max]
	}

	out := map[string]Quote{}
	miss := make([]string, 0, len(syms))
	now := time.Now().UTC()

	for _, sym := range syms {
		if v, ok := s.store.Get(sym); ok {
			out[sym] = v
		} else {
			miss = append(miss, sym)
		}
	}

	if len(miss) > 0 {
		ctxT, cancel := context.WithTimeout(r.Context(), s.cg.Poller.Timeout)
		defer cancel()
		if quotes, err := s.cg.Quotes(ctxT, miss, s.vs); err == nil {
			for sym, price := range quotes {
				if price <= 0 {
					continue
				}
				q := Quote{Symbol: sym, Price: price, Source: s.source, At: now}
				s.store.Set(q)
				out[sym] = q
			}
		}
	}

	hit, missLenght := 0, len(miss)

	w.Header().Set("X-Cache-Hits", strconv.Itoa(hit))
	w.Header().Set("X-Cache-Miss", strconv.Itoa(missLenght))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Source", s.source)
	w.Header().Set("X-TTL-Seconds", strconv.Itoa(int(s.store.ttl.Seconds())))
	// opcional: refresco programado (hint para clientes)
	w.Header().Set("X-Refresh-Every", s.refresh.String())

	_ = json.NewEncoder(w).Encode(out)
}
