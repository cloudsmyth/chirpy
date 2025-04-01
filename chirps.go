package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) getChirpByIdHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameter struct {
		ChirpId uuid.UUID `json:"id"`
	}

	type chirpResponse struct {
		Chirp
	}

	chirpIdString := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad ChirpId used", err)
	}

	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not get chirp from db", err)
		return
	}

	respondWithJson(w, http.StatusOK, chirpResponse{
		Chirp: Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
		},
	})
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	response := []Chirp{}
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirps from db", err)
		return
	}

	for _, chirp := range chirps {
		response = append(response, Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
		})
	}

	respondWithJson(w, http.StatusOK, response)
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type chirpParameters struct {
		Body string `json:"body"`
	}
	type chirpResponse struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to process request header", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}

	words := strings.Split(params.Body, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if stringInMap(lower, banned) {
			words[i] = "****"
		}
	}
	msg := strings.Join(words, " ")

	arg := database.CreateChirpParams{
		Body:   msg,
		UserID: userID,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create new user", err)
		return
	}

	respondWithJson(w, http.StatusCreated, chirpResponse{
		Chirp: Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
		},
	})
}
