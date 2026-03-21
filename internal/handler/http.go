package handler

import (
	"URLshortener/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type ShortenerHandler struct {
	service shortenerService
}

func NewShortenerHandler(service shortenerService) *ShortenerHandler {
	return &ShortenerHandler{service: service}
}

func (h *ShortenerHandler) SaveOrigin(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var req SaveOriginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	shortUrl, err := h.service.AddOrigin(ctx, req.URL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGenerateShortUrl):
			writeError(w, http.StatusInternalServerError, "failed to generate short url")
			return
		case errors.Is(err, service.ErrPersistLink):
			writeError(w, http.StatusInternalServerError, "failed to save short link")
			return
		case errors.Is(err, context.DeadlineExceeded):
			writeError(w, http.StatusGatewayTimeout, "request timeout")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	resp := SaveOriginResponse{Short: shortUrl}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *ShortenerHandler) GetOrigin(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	shortUrl := chi.URLParam(r, "short")

	if shortUrl == "" {
		writeError(w, http.StatusBadRequest, "short url is required")
		return
	}

	originUrl, err := h.service.GetOrigin(ctx, shortUrl)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidShortURL):
			writeError(w, http.StatusBadRequest, "short url must be 10 chars")
			return
		case errors.Is(err, service.ErrNotFound):
			writeError(w, http.StatusNotFound, "origin url not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	resp := GetOriginResponse{URL: originUrl}

	writeJSON(w, http.StatusOK, resp)

}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})

}
