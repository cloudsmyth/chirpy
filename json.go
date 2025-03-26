package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

var banned = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

type responseBody struct {
	Msg   string `json:"error,omitempty"`
	Clean string `json:"cleaned_body"`
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type requestBody struct {
		Body string `json:"body"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	params := requestBody{}
	err = json.Unmarshal(dat, &params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
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
	respondWithJson(w, http.StatusOK, responseBody{Clean: msg})
}

func respondWithJson(w http.ResponseWriter, code int, resp responseBody) error {
	response, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJson(w, code, responseBody{Msg: msg})
}

func stringInMap(s string, m map[string]bool) bool {
	_, exists := m[s]
	return exists
}
