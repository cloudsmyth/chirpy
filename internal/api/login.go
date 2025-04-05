package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/cloudsmyth/chirpy/internal/database"
)

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := cfg.DbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Could not get user with that email", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.Secret)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Error making jwt", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	refresh, err := cfg.DbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})

	err = auth.CheckHashedPassword(user.HashedPassword, params.Password)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	} else {
		common.RespondWithJson(w, http.StatusOK, UserResponse{
			User: User{
				ID:           user.ID,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
				Email:        user.Email,
				Token:        token,
				RefreshToken: refresh.Token,
				IsChirpyRed:  user.IsChirpyRed,
			},
		})
	}
}
