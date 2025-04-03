package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloudsmyth/chirpy/internal/auth"
	"github.com/cloudsmyth/chirpy/internal/common"
)

func (cfg *ApiConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	type refreshResponse struct {
		Token string `json:"token"`
	}

	authHeader, err := auth.GetBearerToken(r.Header)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not get Bearer", err)
		return
	}

	refreshQuery, err := cfg.DbQueries.GetRefreshByToken(r.Context(), authHeader)
	if err != nil {
		common.RespondWithError(w, http.StatusUnauthorized, "Could not find record", err)
		return
	}

	if time.Now().After(refreshQuery.ExpiresAt) {
		common.RespondWithError(w, http.StatusUnauthorized, "Token expired", fmt.Errorf("Refresh token has expired"))
		return
	}

	if refreshQuery.RevokedAt.Valid {
		common.RespondWithError(w, http.StatusUnauthorized, "Token revoked", fmt.Errorf("Refresh token has been revoked"))
		return
	}

	jwtToken, err := auth.MakeJWT(refreshQuery.UserID, cfg.Secret)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Token could not be created", err)
		return
	}

	common.RespondWithJson(w, http.StatusOK, refreshResponse{Token: jwtToken})
}

func (cfg *ApiConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	type revokeResponse struct{}

	authHeader, err := auth.GetBearerToken(r.Header)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not get Bearer", err)
		return
	}

	_, err = cfg.DbQueries.RevokeRefreshByToken(r.Context(), authHeader)
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Refresh was not revoked", err)
		return
	}

	common.RespondWithJson(w, http.StatusNoContent, revokeResponse{})
}
