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
	}

	serverMux := http.NewServeMux()

	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serverMux.HandleFunc("GET /api/healthz", handleReadiness)

	serverMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	serverMux.HandleFunc("POST /api/users", handleCreateNewUser(&apiCfg))

	serverMux.HandleFunc("GET /admin/metrics", handleMetrics(&apiCfg))

	serverMux.HandleFunc("POST /admin/reset", handleReset(&apiCfg))


	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	server.ListenAndServe()

}