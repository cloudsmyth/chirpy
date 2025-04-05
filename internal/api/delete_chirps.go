package api

import (
	"net/http"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/cloudsmyth/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) DeleteChirpsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type response struct{}

	chirpIdString := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdString)
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "Bad ChirpId used", err)
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

	chirp, err := cfg.DbQueries.GetChirpById(r.Context(), chirpId)
	if err != nil {
		common.RespondWithError(w, http.StatusNotFound, "Chrip not found", err)
		return
	}

	if chirp.UserID != validUserId {
		common.RespondWithError(w, http.StatusForbidden, "Can not delete chirp", err)
		return
	}

	if err = cfg.DbQueries.DeleteChirpById(r.Context(), database.DeleteChirpByIdParams{
		UserID: validUserId,
		ID:     chirpId,
	}); err != nil {
		common.RespondWithError(w, http.StatusForbidden, "Could not delete chirp", err)
		return
	}

	common.RespondWithJson(w, http.StatusNoContent, response{})
}
