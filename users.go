package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type response struct {
	User
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not get user with that email", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making jwt", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	refresh, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})

	err = auth.CheckHashedPassword(user.HashedPassword, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	} else {
		respondWithJson(w, http.StatusOK, response{
			User: User{
				ID:           user.ID,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
				Email:        user.Email,
				Token:        token,
				RefreshToken: refresh.Token,
			},
		})
	}
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	type refreshResponse struct {
		Token string `json:"token"`
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Missing auth token", fmt.Errorf("No Authorization token found"))
		return
	}

	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "Malformed Auth token", fmt.Errorf("Malformed Auth token"))
		return
	}

	refreshQuery, err := cfg.dbQueries.GetRefreshByToken(r.Context(), splitAuth[1])
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not find record", err)
		return
	}

	if time.Now().After(refreshQuery.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", err)
		return
	}

	if refreshQuery.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", err)
		return
	}

	jwtToken, err := auth.MakeJWT(refreshQuery.UserID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Token could not be created", err)
		return
	}

	respondWithJson(w, http.StatusOK, refreshResponse{Token: jwtToken})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	type revokeResponse struct{}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusBadRequest, "Missing auth token", fmt.Errorf("No Authorization token found"))
		return
	}

	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		respondWithError(w, http.StatusBadRequest, "Malformed Auth token", fmt.Errorf("Malformed Auth token"))
		return
	}

	_, err := cfg.dbQueries.RevokeRefreshByToken(r.Context(), splitAuth[1])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh was not revoked", err)
		return
	}

	respondWithJson(w, http.StatusNoContent, revokeResponse{})
}

func (cfg *apiConfig) addUserHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	arg := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), arg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create new user", err)
		return
	}

	respondWithJson(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}
