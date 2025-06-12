package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type handler func(http.ResponseWriter, *http.Request)

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK\n"))
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct{
		Body string `json:"body"`
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
	respondWithJSONResponseBody(w, cleanedBody)
}

func handleCreateNewUser(apiCfg *apiConfig) handler {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
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
		
		user, err := apiCfg.db.CreateUser(r.Context(), params.Email)
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
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		})
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
	type errorResponseBody struct{
		Error string `json:"error"`
	}

	resp_data, resp_err := json.Marshal(
		errorResponseBody{
			Error: errorString,
		},
	)
	if resp_err != nil {
		log.Printf("error marshalling error response body: %v", resp_err)
		w.WriteHeader(500)
		return
	}
	w.Write(resp_data)
}

func respondWithJSONResponseBody(w http.ResponseWriter, cleanedBodyString string) {
	type cleanedResponseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	cleanedRespBody := cleanedResponseBody{
		CleanedBody: cleanedBodyString,
	}
	payload, err := json.Marshal(cleanedRespBody)
	if err != nil {
		log.Printf("error marshalling cleaned response body: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(payload)
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