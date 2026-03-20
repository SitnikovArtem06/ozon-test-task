package handler

import (
	"URLshortener/internal/service"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type ShortenerHandler struct {
	service shortenerService
}

func NewShortenerHandler(service shortenerService) *ShortenerHandler {
	return &ShortenerHandler{service: service}
}

func (h *ShortenerHandler) SaveOrigin(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req SaveOriginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	shortUrl, err := h.service.AddOrigin(ctx, req.URL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGenerateShortUrl):
			http.Error(w, "failed to generate short url", http.StatusInternalServerError)
			return
		case errors.Is(err, service.ErrPersistLink):
			http.Error(w, "failed to save short link", http.StatusInternalServerError)
			return
		case errors.Is(err, context.DeadlineExceeded):
			http.Error(w, "request timeout", http.StatusGatewayTimeout)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	resp := SaveOriginResponse{Short: shortUrl}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *ShortenerHandler) GetOrigin(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	shortUrl := chi.URLParam(r, "short")

	if shortUrl == "" {
		http.Error(w, "short url is required", http.StatusBadRequest)
		return
	}

	originUrl, err := h.service.GetOrigin(ctx, shortUrl)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidShortURL):
			http.Error(w, "short url must be 10 chars", http.StatusBadRequest)
			return
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "origin url not found", http.StatusNotFound)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	resp := GetOriginResponse{URL: originUrl}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

}
