package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func respondWithJson(w http.ResponseWriter, code int, resp interface{}) {
	response, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Printf("Error: %s\n", err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJson(w, code, errorResponse{
		Error: msg,
	})
}

func stringInMap(s string, m map[string]bool) bool {
	_, exists := m[s]
	return exists
}
