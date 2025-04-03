package api

import (
	"net/http"

	"github.com/cloudsmyth/chirpy/internal/common"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	response := []Chirp{}
	chirps, err := cfg.DbQueries.GetChirps(r.Context())
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "Could not get chirps from db", err)
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

	common.RespondWithJson(w, http.StatusOK, response)
}

func (cfg *ApiConfig) GetChirpByIdHandler(w http.ResponseWriter, r *http.Request) {
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
		common.RespondWithError(w, http.StatusBadRequest, "Bad ChirpId used", err)
	}

	chirp, err := cfg.DbQueries.GetChirpById(r.Context(), chirpId)
	if err != nil {
		common.RespondWithError(w, http.StatusNotFound, "Could not get chirp from db", err)
		return
	}

	common.RespondWithJson(w, http.StatusOK, chirpResponse{
		Chirp: Chirp{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserId:    chirp.UserID,
		},
	})
}
