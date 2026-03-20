package service_test

import (
	"URLshortener/internal/repository"
	"URLshortener/internal/service"
	"URLshortener/internal/service/mocks"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
)

func runTx(t *testing.T, tm *mocks.MockTransactionManager, withTx bool) {
	t.Helper()
	tm.EXPECT().
		Do(gomock.Any(), withTx, gomock.Any()).
		DoAndReturn(func(parent context.Context, txFlag bool, fn func(context.Context) error) error {
			if txFlag != withTx {
				t.Fatalf("unexpected withTx flag: got %v want %v", txFlag, withTx)
			}
			return fn(parent)
		})
}

func TestAddOrigin_ReturnsExistingShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)
	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("abc123_DEF", nil)

	got, err := svc.AddOrigin(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123_DEF" {
		t.Fatalf("unexpected short: got %q", got)
	}
}

func TestAddOrigin_CreatesNewShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)

	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", repository.ErrNotFoundOrigin)
	repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", repository.ErrNotFoundShort)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), "https://example.com").Return(nil)

	_, err := svc.AddOrigin(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}

func TestAddOrigin_RetriesOnDuplicateShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)

	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", repository.ErrNotFoundOrigin)
	repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", repository.ErrNotFoundShort)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), "https://example.com").Return(repository.ErrDuplicateShort)
	repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", repository.ErrNotFoundShort)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), "https://example.com").Return(nil)

	_, err := svc.AddOrigin(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}

func TestAddOrigin_ReturnsExistingAfterDuplicateOrigin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)

	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", repository.ErrNotFoundOrigin)
	repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", repository.ErrNotFoundShort)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), "https://example.com").Return(repository.ErrDuplicateOrigin)
	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("abc123_DEF", nil)

	got, err := svc.AddOrigin(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123_DEF" {
		t.Fatalf("unexpected short: got %q", got)
	}
}

func TestAddOrigin_RepoErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)
	dbErr := errors.New("db error")
	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", dbErr)

	_, err := svc.AddOrigin(context.Background(), "https://example.com")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected db error, got: %v", err)
	}
}

func TestAddOrigin_ReturnsGetErrAfterDuplicateOrigin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)
	getErr := errors.New("failed to get existing short")

	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", repository.ErrNotFoundOrigin)
	repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", repository.ErrNotFoundShort)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), "https://example.com").Return(repository.ErrDuplicateOrigin)
	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", getErr)

	_, err := svc.AddOrigin(context.Background(), "https://example.com")
	if !errors.Is(err, getErr) {
		t.Fatalf("expected getErr, got: %v", err)
	}
}

func TestAddOrigin_ReturnsPersistErrorOnCreateDefaultBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)
	createErr := errors.New("insert failed")

	repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", repository.ErrNotFoundOrigin)
	repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", repository.ErrNotFoundShort)
	repo.EXPECT().Create(gomock.Any(), gomock.Any(), "https://example.com").Return(createErr)

	_, err := svc.AddOrigin(context.Background(), "https://example.com")
	if !errors.Is(err, service.ErrPersistLink) {
		t.Fatalf("expected ErrPersistLink, got: %v", err)
	}
}

func TestAddOrigin_GetByShortError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, true)
	dbErr := errors.New("get by short failed")

	gomock.InOrder(
		repo.EXPECT().GetByOrigin(gomock.Any(), "https://example.com").Return("", repository.ErrNotFoundOrigin),
		repo.EXPECT().GetByShort(gomock.Any(), gomock.Any()).Return("", dbErr),
	)

	_, err := svc.AddOrigin(context.Background(), "https://example.com")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected get-by-short error, got: %v", err)
	}
}

func TestGetOrigin_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, false)
	repo.EXPECT().GetByShort(gomock.Any(), "abc123_DEF").Return("", repository.ErrNotFoundShort)

	_, err := svc.GetOrigin(context.Background(), "abc123_DEF")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetOrigin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, false)
	repo.EXPECT().GetByShort(gomock.Any(), "abc123_DEF").Return("https://example.com", nil)

	got, err := svc.GetOrigin(context.Background(), "abc123_DEF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.com" {
		t.Fatalf("unexpected origin: got %q", got)
	}
}

func TestGetOrigin_InvalidShortLength(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	_, err := svc.GetOrigin(context.Background(), "short")
	if !errors.Is(err, service.ErrInvalidShortURL) {
		t.Fatalf("expected ErrInvalidShortURL, got: %v", err)
	}
}

func TestGetOrigin_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	tm := mocks.NewMockTransactionManager(ctrl)
	svc := service.NewShortenerService(tm, repo)

	runTx(t, tm, false)
	dbErr := errors.New("db error")
	repo.EXPECT().GetByShort(gomock.Any(), "abc123_DEF").Return("", dbErr)

	_, err := svc.GetOrigin(context.Background(), "abc123_DEF")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
