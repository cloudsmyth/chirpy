package api

import (
	"encoding/json"
	"net/http"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/cloudsmyth/chirpy/internal/database"
)

func (cfg *ApiConfig) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	arg := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	user, err := cfg.DbQueries.CreateUser(r.Context(), arg)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not create new user", err)
		return
	}

	common.RespondWithJson(w, http.StatusCreated, UserResponse{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
