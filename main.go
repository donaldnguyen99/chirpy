package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"

	"github.com/donaldnguyen99/chirpy/internal/database"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	tokenSecret string
	polkaKey string
	accessExpiresInSeconds int
	refreshExpiresInHours int
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

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db: database.New(db),
		platform: os.Getenv("PLATFORM"),
		tokenSecret: os.Getenv("TOKEN_SECRET"),
		polkaKey: os.Getenv("POLKA_KEY"),
		accessExpiresInSeconds: 3600,
		refreshExpiresInHours: 24 * 60,
	}

	serverMux := http.NewServeMux()

	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serverMux.HandleFunc("GET /api/healthz", handleReadiness)
	serverMux.HandleFunc("GET /api/chirps", handleGetAllChirps(&apiCfg))
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", handleGetChirpByID(&apiCfg))
	serverMux.HandleFunc("POST /api/chirps", handleNewChirp(&apiCfg))
	serverMux.HandleFunc("DELETE /api/chirps/{chirpID}", handleDeleteChirpByID(&apiCfg))
	serverMux.HandleFunc("POST /api/login", handleLogin(&apiCfg))
	serverMux.HandleFunc("POST /api/refresh", handleRefresh(&apiCfg))
	serverMux.HandleFunc("POST /api/revoke", handleRevoke(&apiCfg))
	serverMux.HandleFunc("POST /api/users", handleCreateNewUser(&apiCfg))
	serverMux.HandleFunc("PUT /api/users", handleUpdateUser(&apiCfg))
	serverMux.HandleFunc("POST /api/polka/webhooks", handleUpgradeUserToChirpyRed(&apiCfg))

	serverMux.HandleFunc("GET /admin/metrics", handleMetrics(&apiCfg))

	serverMux.HandleFunc("POST /admin/reset", handleReset(&apiCfg))


	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	server.ListenAndServe()

}