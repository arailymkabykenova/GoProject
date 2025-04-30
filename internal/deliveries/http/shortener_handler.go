package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"template/internal/repositories"
	"template/internal/services"
	models "template/internal/usecases/shortner"
)

type UpdateRequest struct {
	NewURL string `json:"new_url"`
}

type ShortenerHandler struct {
	service services.ShortenerService
	repo    repositories.ShortenerRepository
	baseURL string
}

func NewShortenerHandler(svc services.ShortenerService, repo repositories.ShortenerRepository, baseURL string) *ShortenerHandler {
	return &ShortenerHandler{
		service: svc,
		repo:    repo,
		baseURL: baseURL,
	}
}

func (h *ShortenerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/shorten", h.handleShorten)
	mux.HandleFunc("/update/", h.handleUpdate)
	mux.HandleFunc("/delete/", h.handleDelete)
	mux.HandleFunc("/", h.handleRedirectOrRoot)

	log.Println("Shortener routes registered: POST /shorten, PUT /update/, DELETE /delete/, GET /")
}

func (h *ShortenerHandler) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Handler error decoding shorten request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	shortCode, err := h.service.CreateShortURL(req.URL)
	if err != nil {
		log.Printf("Handler error from service CreateShortURL: %v", err)
		if strings.Contains(err.Error(), "invalid URL format") {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to create short URL")
		}
		return
	}

	fullShortURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(h.baseURL, "/"), shortCode)
	resp := models.ShortenResponse{ShortURL: fullShortURL, OriginalURL: req.URL}
	respondWithJSON(w, http.StatusCreated, resp)
	log.Printf("Handler successfully handled shorten request for %s -> %s", req.URL, fullShortURL)
}

func (h *ShortenerHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/update/")
	if shortCode == "" || strings.Contains(shortCode, "/") {
		respondWithError(w, http.StatusBadRequest, "Invalid short code in URL path")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Handler error decoding update request for code %s: %v", shortCode, err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer r.Body.Close()

	if req.NewURL == "" {
		respondWithError(w, http.StatusBadRequest, "Missing 'new_url' in request body")
		return
	}

	err := h.service.UpdateLongURL(shortCode, req.NewURL)
	if err != nil {
		log.Printf("Handler error from service UpdateLongURL for code %s: %v", shortCode, err)
		if errors.Is(err, repositories.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Short code not found")
		} else if strings.Contains(err.Error(), "invalid new URL format") {
			respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update mapping")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "URL updated successfully"})
	log.Printf("Handler successfully updated short code %s", shortCode)
}

func (h *ShortenerHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/delete/")
	if shortCode == "" || strings.Contains(shortCode, "/") {
		respondWithError(w, http.StatusBadRequest, "Invalid short code in URL path")
		return
	}

	err := h.service.DeleteMapping(shortCode)
	if err != nil {
		log.Printf("Handler error from service DeleteMapping for code %s: %v", shortCode, err)
		if errors.Is(err, repositories.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Short code not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete mapping")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Handler successfully deleted short code %s", shortCode)
}

func (h *ShortenerHandler) handleRedirectOrRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	if r.URL.Path == "/" {
		respondWithJSON(w, http.StatusOK, map[string]string{"message": "URL Shortener API. Use POST /shorten, PUT /update/{code}, DELETE /delete/{code}, or GET /{code}"})
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/update/") || strings.HasPrefix(r.URL.Path, "/delete/") {
		http.NotFound(w, r)
		return
	}

	longURL, err := h.repo.FindByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			log.Printf("Handler: Short code not found: %s", shortCode)
			respondWithError(w, http.StatusNotFound, "Short code not found")
		} else {
			log.Printf("Handler: Database error during redirect lookup for code %s: %v", shortCode, err)
			respondWithError(w, http.StatusInternalServerError, "Error looking up short code")
		}
		return
	}

	log.Printf("Handler: Redirecting code %s to %s", shortCode, longURL)
	http.Redirect(w, r, longURL, http.StatusFound)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	log.Printf("Responding with error: %d - %s", code, message)
	respondWithJSON(w, code, models.ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error marshalling response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}
