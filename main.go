package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetMetrics() {
	cfg.fileserverHits.Store(0)
}

func main() {

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	serverMux := http.NewServeMux()

	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serverMux.HandleFunc("/healthz", handleReadiness)

	serverMux.HandleFunc("/metrics", handleMetrics(&apiCfg))

	serverMux.HandleFunc("/reset", handleResetMetrics(&apiCfg))

	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	server.ListenAndServe()

}