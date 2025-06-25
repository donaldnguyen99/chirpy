package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/donaldnguyen99/chirpy/internal/auth"
	"github.com/donaldnguyen99/chirpy/internal/database"
	"github.com/google/uuid"
)

type handler func(http.ResponseWriter, *http.Request)

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK\n"))
}

func handleGetAllChirps(apiCgf *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		chirps, err := apiCgf.db.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("error querying for all chirps: %v", err)
			w.WriteHeader(500)
			return
		}

		type chirpResponse struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    uuid.UUID `json:"user_id"`
		}

		chirpRespSlice := []chirpResponse{}
		for _, chirp := range chirps {
			chirpResp := chirpResponse{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			}
			chirpRespSlice = append(chirpRespSlice, chirpResp)
		}
		payload, err := json.Marshal(chirpRespSlice)
		payload = append(payload, "\n"...)
		if err != nil {
			log.Printf("error marshalling chirp response: %v", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(payload)
	}
}

func handleGetChirpByID(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		chirpID := r.PathValue("chirpID")
		chirpUUID, err := uuid.Parse(chirpID)
		if err != nil {
			w.WriteHeader(404)
			log.Printf("error parsing id: %v", err)
			respondWithErrorResponseBody(w, "Not found")
			return
		}

		chirp, err := apiCfg.db.GetChirpByID(r.Context(), chirpUUID)
		if err != nil {
			w.WriteHeader(404)
			log.Printf("error querying chirp by id: %v", err)
			respondWithErrorResponseBody(w, "Not found")
			return
		}

		type chirpResponse struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    uuid.UUID `json:"user_id"`
		}

		chirpResp := chirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		payload, err := json.Marshal(chirpResp)
		payload = append(payload, "\n"...)
		if err != nil {
			log.Printf("error marshalling chirp response: %v", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		w.Write(payload)
	}
}

func handleNewChirp(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body   string    `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("error decoding parameters: %v", err)
			w.WriteHeader(500)
			return
		}

		if params.Body == "" {
			const errorString = "Request body is empty"
			log.Printf("error: " + errorString)
			w.WriteHeader(400)
			respondWithErrorResponseBody(w, errorString)
			return
		}

		if len(params.Body) > 140 {
			const errorString = "Chirp is too long"
			log.Printf("error: " + errorString)
			w.WriteHeader(400)
			respondWithErrorResponseBody(w, errorString)
			return
		}

		cleanedBody := replaceProfanity(params.Body)

		jwt, err := auth.GetBearerToken(r.Header)
		if err != nil {
			if strings.Contains(err.Error(), "'Bearer ' prefix") {
				w.WriteHeader(400)
				respondWithErrorResponseBody(w, err.Error())
			} else {
				log.Printf("error parsing bearer token while making new chirp: %v", err)
				w.WriteHeader(500)
			}
			return
		}

		resUUID, err := auth.ValidateJWT(jwt, apiCfg.tokenSecret)
		if err != nil {
			w.WriteHeader(401)
			respondWithErrorResponseBody(w, "Invalid token")
			return
		}

		chirp, err := apiCfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: resUUID,
		})

		if err != nil {
			log.Printf("error creating chirp: %v", err)
			w.WriteHeader(500)
			return
		}
		log.Printf("Chirp created for %v", chirp.UserID)

		type chirpResponse struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    uuid.UUID `json:"user_id"`
		}

		chirpResp := chirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		payload, err := json.Marshal(chirpResp)
		payload = append(payload, "\n"...)
		if err != nil {
			log.Printf("error marshalling chirp response: %v", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(201)
		w.Write(payload)
	}
}

func handleLogin(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Password string `json:"password"`
			Email string `json:"email"`
			ExpiresInSeconds int `json:"expires_in_seconds"`
		}
		params := parameters{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("error decoding parameters: %v", err)
			w.WriteHeader(500)
			return
		}

		if params.ExpiresInSeconds <= 0 || params.ExpiresInSeconds > 3600 {
			params.ExpiresInSeconds = 3600
		}

		user, err := apiCfg.db.GetUser(r.Context(), params.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Print("login attempted, user not found")
				w.WriteHeader(401)
				respondWithErrorResponseBody(w, "Incorrect email or password")
			} else {
				log.Printf("error getting user: %v", err)
				w.WriteHeader(500)
			}
			return
		}
		
		if err = auth.CheckPasswordHash(user.HashedPassword, params.Password); err != nil {
			log.Printf("login attempted for user %s password incorrect: %v", user.Email, err)
			w.WriteHeader(401)
			respondWithErrorResponseBody(w, "Incorrect email or password")
			return
		}
		
		log.Printf("User %s successfully logged in", user.Email)

		jwt, err := auth.MakeJWT(user.ID, apiCfg.tokenSecret, time.Duration(params.ExpiresInSeconds) * time.Second)
		if err != nil {
			log.Printf("error making jwt: %v", err)
			w.WriteHeader(500)
			return
		}

		type User struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
			Token	  string	`json:"token"`
		}
		data, err := json.Marshal(User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     jwt,
		})

		data = append(data, "\n"...)
		if err != nil {
			log.Printf("error marshalling user: %v", err)
			w.WriteHeader(500)
		}
		w.WriteHeader(200)
		w.Write(data)
	}
}

func handleCreateNewUser(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Password string `json:"password"`
			Email string `json:"email"`
		}
		params := parameters{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("error decoding parameters: %v", err)
			w.WriteHeader(500)
			return
		}

		hashed_password, err := auth.HashPassword(params.Password)
		if err != nil {
			log.Printf("error hashing password for new user: %v", err)
			w.WriteHeader(500)
			return
		}

		user, err := apiCfg.db.CreateUser(r.Context(), database.CreateUserParams{
			HashedPassword: hashed_password,
			Email: params.Email,
		})
		if err != nil {
			log.Printf("error creating user: %v", err)
			w.WriteHeader(500)
			return
		}
		log.Printf("User created successfully: %s", user.Email)

		type User struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
		}
		data, err := json.Marshal(User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		})

		data = append(data, "\n"...)
		if err != nil {
			log.Printf("error marshalling user: %v", err)
			w.WriteHeader(500)
		}
		w.WriteHeader(201)
		w.Write(data)
	}
}

func handleMetrics(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(
			fmt.Sprintf(
				`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, apiCfg.fileserverHits.Load()),
		))
	}
}

func handleReset(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiCfg.platform == "dev" {
			err := apiCfg.db.DeleteAllUsers(r.Context())
			if err != nil {
				log.Printf("error deleting all users: %v", err)
				w.WriteHeader(500)
			}

			apiCfg.resetMetrics()
			r.Header.Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(200)
			w.Write([]byte("Server reset\n"))
			log.Println("Server reset")
		} else {
			log.Println("Request submitted not as dev")
			w.WriteHeader(403)
			w.Write([]byte("403 Forbidden\n"))
		}
	}
}

func respondWithErrorResponseBody(w http.ResponseWriter, errorString string) {
	type errorResponseBody struct {
		Error string `json:"error"`
	}

	resp_data, resp_err := json.Marshal(
		errorResponseBody{
			Error: errorString,
		},
	)
	resp_data = append(resp_data, "\n"...)
	if resp_err != nil {
		log.Printf("error marshalling error response body: %v", resp_err)
		w.WriteHeader(500)
		return
	}
	w.Write(resp_data)
}

func replaceProfanity(s string) string {
	words := strings.Split(s, " ")
	lower_words := strings.Split(strings.ToLower(s), " ")
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	for i, word := range lower_words {
		if slices.Contains(profaneWords, word) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
