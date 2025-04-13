package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/cloudsmyth/chirpy/internal/database"
	"github.com/google/uuid"
)

type webhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *ApiConfig) UpgradeChirpyRedHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Could not get api key", err)
		return
	}

	if apiKey != cfg.Polka {
		common.RespondWithError(w, http.StatusUnauthorized, "Incorrect API key", errors.New("Incorrect api key"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := webhookRequest{}
	if err := decoder.Decode(&params); err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		common.RespondWithJson(w, http.StatusNoContent, UserResponse{})
		return
	}

	userUUID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not parse UUID", err)
		return
	}

	user, err := cfg.DbQueries.UpgradeUserById(r.Context(), database.UpgradeUserByIdParams{
		IsChirpyRed: true,
		ID:          userUUID,
	})
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}

	if user.ID == uuid.Nil {
		common.RespondWithError(w, http.StatusNotFound, "Could not find user", err)
		return
	}

	common.RespondWithJson(w, http.StatusNoContent, UserResponse{})
}
