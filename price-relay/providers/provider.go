package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/ownhttp"
)

type ProviderClient struct {
	Poller *ownhttp.Poller
	Base   string
	UA     string
	source string
	Tpl    string // SOURCE_URL_TPL (con {base}, {ids}, {vs})

}

func NewProviderClient(timeout time.Duration) *ProviderClient {
	p := ownhttp.NewPoller(443, timeout, 1, nil)
	c := &ProviderClient{
		Poller: p,
		Base:   helpers.GetEnvOrFail("SOURCE_URL"),
		UA:     "MoonMap-PriceRelay/1.0",
	}

	c.source = helpers.GetEnvOrFail("SOURCE")
	c.Tpl = helpers.GetEnvOrFail("SOURCE_URL_TPL")

	return c
}

func (c *ProviderClient) buildURL(ids []string, vsCurrencies []string) (string, error) {
	if len(ids) == 0 || len(vsCurrencies) == 0 {
		return "", errors.New("missing ids or vs")
	}
	idsLower := strings.ToLower(strings.Join(ids, ","))
	vsLower := strings.ToLower(strings.Join(vsCurrencies, ","))

	// encode solo los valores, no toda la URL
	idsEnc := url.QueryEscape(idsLower)
	vsEnc := url.QueryEscape(vsLower)

	u := c.Tpl
	u = strings.ReplaceAll(u, "{base}", strings.TrimRight(c.Base, "/"))
	u = strings.ReplaceAll(u, "{ids}", idsEnc)
	u = strings.ReplaceAll(u, "{vs}", vsEnc)

	return u, nil
}

func (c *ProviderClient) SimplePrice(ctx context.Context, ids []string, vsCurrencies []string) (map[string]map[string]float64, error) {

	u, err := c.buildURL(ids, vsCurrencies)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	h := http.Header{}
	h.Set("Accept", "application/json")
	if c.UA != "" {
		h.Set("User-Agent", c.UA)
	}
	c.Poller.SetHeaders(req, h)

	res, err := c.Poller.Httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	defer c.Poller.Httpc.CloseIdleConnections()

	if res.StatusCode == 429 {
		if ra := res.Header.Get("Retry-After"); ra != "" {
			if n, _ := strconv.Atoi(ra); n > 0 {
				time.Sleep(time.Duration(n) * time.Second)
			}
		}

		return nil, fmt.Errorf("%s 429", c.source)
	}
	if res.StatusCode >= 500 {
		return nil, fmt.Errorf("%s %d", c.source, res.StatusCode)
	}

	var out map[string]map[string]float64
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	logrus.Info("new prices fetched from coin gecko")
	return out, nil
}

var DefaultSymbolMap = map[string]string{
	"SOL": "solana",
	"BTC": "bitcoin",
	"ETH": "ethereum",
}

func (c *ProviderClient) Quotes(ctx context.Context, symbols []string, vs string) (map[string]float64, error) {
	idx := map[string]string{}
	ids := make([]string, 0, len(symbols))
	for _, s := range symbols {
		up := strings.ToUpper(s)
		id := DefaultSymbolMap[up]
		if id == "" {
			continue
		}
		idx[id] = up
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("no supported symbols")
	}
	data, err := c.SimplePrice(ctx, ids, []string{vs})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	out := make(map[string]float64, len(symbols))
	vsL := strings.ToLower(vs)
	for id, m := range data {
		if sym := idx[id]; sym != "" {
			out[sym] = m[vsL]
		}
	}
	return out, nil
}
