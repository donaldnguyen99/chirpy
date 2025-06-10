package main

import (
	"fmt"
	"net/http"
)

type handler func(http.ResponseWriter, *http.Request)

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK\n"))
}

func handleMetrics(apiCfg *apiConfig) (handler) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(
			fmt.Sprintf("Hits: %v\n", apiCfg.fileserverHits.Load()),
		))
	}
}

func handleResetMetrics(apiCfg *apiConfig) (handler) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiCfg.resetMetrics()
		r.Header.Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("Hits reset\n"))
	}
}