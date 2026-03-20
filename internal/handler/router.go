package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *ShortenerHandler) http.Handler {
	r := chi.NewRouter()

	r.Post("/shorten", h.SaveOrigin)
	r.Get("/{short}", h.GetOrigin)

	return r
}
