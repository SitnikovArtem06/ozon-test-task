package handler

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func NewRouter(h *ShortenerHandler) http.Handler {
	r := chi.NewRouter()

	r.Post("/shorten", h.SaveOrigin)
	r.Get("/{short}", h.GetOrigin)

	return r
}
