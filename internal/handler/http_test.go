package handler_test

import (
	"URLshortener/internal/handler"
	"URLshortener/internal/handler/mocks"
	"URLshortener/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestSaveOrigin_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("{invalid"))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestSaveOrigin_EmptyURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":""}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestSaveOrigin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		AddOrigin(gomock.Any(), "https://example.com").
		DoAndReturn(func(ctx context.Context, origin string) (string, error) {
			return "abc123_DEF", nil
		})

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp["short_url"] != "abc123_DEF" {
		t.Fatalf("unexpected short_url: %q", resp["short_url"])
	}
}

func TestSaveOrigin_PersistError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		AddOrigin(gomock.Any(), "https://example.com").
		Return("", service.ErrPersistLink)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestSaveOrigin_GenerateShortError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		AddOrigin(gomock.Any(), "https://example.com").
		Return("", service.ErrGenerateShortUrl)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestSaveOrigin_RequestTimeoutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		AddOrigin(gomock.Any(), "https://example.com").
		Return("", context.DeadlineExceeded)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status 504, got %d", rec.Code)
	}
}

func TestSaveOrigin_UnknownError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		AddOrigin(gomock.Any(), "https://example.com").
		Return("", errors.New("some error"))

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestGetOrigin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		GetOrigin(gomock.Any(), "abc123_DEF").
		Return("https://example.com", nil)

	req := httptest.NewRequest(http.MethodGet, "/abc123_DEF", bytes.NewReader(nil))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if resp["url"] != "https://example.com" {
		t.Fatalf("unexpected url: %q", resp["url"])
	}
}

func TestGetOrigin_InvalidShortLength(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		GetOrigin(gomock.Any(), "short").
		Return("", service.ErrInvalidShortURL)

	req := httptest.NewRequest(http.MethodGet, "/short", bytes.NewReader(nil))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestGetOrigin_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		GetOrigin(gomock.Any(), "abc123_DEF").
		Return("", service.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/abc123_DEF", bytes.NewReader(nil))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestGetOrigin_UnknownError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockshortenerService(ctrl)
	h := handler.NewShortenerHandler(svc)
	r := handler.NewRouter(h)

	svc.EXPECT().
		GetOrigin(gomock.Any(), "abc123_DEF").
		Return("", errors.New("unexpected"))

	req := httptest.NewRequest(http.MethodGet, "/abc123_DEF", bytes.NewReader(nil))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
