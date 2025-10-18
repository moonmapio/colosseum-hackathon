package ownhttp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

type Poller struct {
	Httpc       *http.Client
	Port        int
	Timeout     time.Duration
	Paths       []string
	Concurrency int
}

func NewPoller(port int, timeout time.Duration, concurrency int, paths []string) *Poller {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Poller{
		Httpc:       &http.Client{Transport: tr, Timeout: timeout},
		Port:        port,
		Timeout:     timeout,
		Paths:       paths,
		Concurrency: concurrency,
	}
}

func (p *Poller) FetchJSON(ctx context.Context, url string, headers http.Header) (interface{}, []byte, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	p.SetHeaders(req, headers)
	res, err := p.Httpc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	var v interface{}
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, nil, err
	}
	return v, nil, nil
}

func (p *Poller) FetchAny(ctx context.Context, url string, headers http.Header) (interface{}, []byte, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	p.SetHeaders(req, headers)
	res, err := p.Httpc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	ct := res.Header.Get("Content-Type")
	if strings.Contains(ct, "json") || strings.HasSuffix(url, ".json") {
		var v interface{}
		dec := json.NewDecoder(res.Body)
		dec.UseNumber()
		if err := dec.Decode(&v); err != nil {
			return nil, nil, err
		}
		return v, nil, nil
	}
	b := make([]byte, 0, 2048)
	tmp := make([]byte, 2048)
	for {
		n, er := res.Body.Read(tmp)
		if n > 0 {
			b = append(b, tmp[:n]...)
		}
		if er != nil {
			break
		}
	}
	return string(b), b, nil
}

func (p *Poller) SetHeaders(req *http.Request, headers http.Header) {
	if headers == nil {
		return
	}

	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
}
