package api

import (
	"encoding/json"
	"net/http"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/cloudsmyth/chirpy/internal/database"
)

func (cfg *ApiConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	authHeader, err := auth.GetBearerToken(r.Header)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Unable to get auth token", err)
		return
	}

	validUserId, err := auth.ValidateJWT(authHeader, cfg.Secret)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Token not correct", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	newUser, err := cfg.DbQueries.UpdateUserById(r.Context(), database.UpdateUserByIdParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             validUserId,
	})
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}

	common.RespondWithJson(w, http.StatusOK, UserResponse{
		User: User{
			ID:          newUser.ID,
			CreatedAt:   newUser.CreatedAt,
			UpdatedAt:   newUser.UpdatedAt,
			Email:       newUser.Email,
			IsChirpyRed: newUser.IsChirpyRed,
		},
	})
}
