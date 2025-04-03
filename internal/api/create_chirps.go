package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/cloudsmyth/chirpy/internal/database"
)

func (cfg *ApiConfig) CreateChirpsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var banned = map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	type chirpParameters struct {
		Body string `json:"body"`
	}

	type chirpResponse struct {
		Chirp
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Unable to process request header", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Unauthorized user", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParameters{}
	if err := decoder.Decode(&params); err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	if len(params.Body) > 140 {
		common.RespondWithError(w, 400, "Chirp is too long", nil)
		return
	}

	words := strings.Split(params.Body, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if common.StringInMap(lower, banned) {
			words[i] = "****"
		}
	}
	msg := strings.Join(words, " ")

	arg := database.CreateChirpParams{
		Body:   msg,
		UserID: userID,
	}

	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), arg)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not create new user", err)
		return
	}

	common.RespondWithJson(w, http.StatusCreated, chirpResponse{
		Chirp: Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
		},
	})
}
