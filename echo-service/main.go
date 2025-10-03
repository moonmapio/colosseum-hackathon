package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

type EchoResponse struct {
	Method   string              `json:"method"`
	URL      string              `json:"url"`
	Proto    string              `json:"proto"`
	Headers  map[string][]string `json:"headers"`
	Query    map[string][]string `json:"query"`
	Body     string              `json:"body,omitempty"`
	Cookies  []map[string]string `json:"cookies"`
	RemoteIP string              `json:"remoteIp"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Dump completo en crudo (útil para debugging hardcore)
	raw, _ := httputil.DumpRequest(r, true)
	log.Println("=== RAW REQUEST ===")
	log.Println(string(raw))
	log.Println("===================")

	// Leer body
	bodyBytes, _ := io.ReadAll(r.Body)

	// Procesar cookies
	var cookies []map[string]string
	for _, c := range r.Cookies() {
		cookies = append(cookies, map[string]string{
			"name":  c.Name,
			"value": c.Value,
		})
	}

	// Respuesta JSON estructurada
	resp := EchoResponse{
		Method:   r.Method,
		URL:      r.URL.String(),
		Proto:    r.Proto,
		Headers:  r.Header,
		Query:    r.URL.Query(),
		Body:     string(bodyBytes),
		Cookies:  cookies,
		RemoteIP: r.RemoteAddr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/", handler)
	port := "8080"
	log.Printf("Echo server with cookies listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
