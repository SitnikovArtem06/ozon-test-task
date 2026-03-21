package service

import (
	"URLshortener/internal/generator"
	"URLshortener/internal/repository"
	"URLshortener/internal/tx"
	"context"
	"errors"
)

type ShortenerService struct {
	txManager tx.TransactionManager
	repo      Repository
}

func NewShortenerService(tm tx.TransactionManager, repo Repository) *ShortenerService {
	return &ShortenerService{txManager: tm, repo: repo}
}

func (s *ShortenerService) AddOrigin(ctx context.Context, originUrl string) (string, error) {

	var resp string
	err := s.txManager.Do(ctx, true, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

		}

		var shortUrl string
		var err error

		shortUrl, err = s.repo.GetByOrigin(ctx, originUrl)
		if err == nil {
			resp = shortUrl
			return nil
		}
		if !errors.Is(err, repository.ErrNotFoundOrigin) {
			return err
		}

		for {
			shortUrl, done, stepErr := s.addOriginStep(ctx, originUrl)
			if stepErr != nil {
				return stepErr
			}
			if done {
				resp = shortUrl
				return nil
			}
		}
	})

	return resp, err
}

func (s *ShortenerService) GetOrigin(ctx context.Context, shortUrl string) (string, error) {
	if len(shortUrl) != 10 {
		return "", ErrInvalidShortURL
	}

	var resp string

	err := s.txManager.Do(ctx, false, func(ctx context.Context) error {

		if originUrl, err := s.repo.GetByShort(ctx, shortUrl); err != nil {
			if errors.Is(err, repository.ErrNotFoundShort) {
				return ErrNotFound
			}
			return err
		} else {
			resp = originUrl
			return nil
		}
	})

	return resp, err

}

func (s *ShortenerService) addOriginStep(ctx context.Context, originUrl string) (shortUrl string, done bool, err error) {
	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", false, ErrGenerateTimeout
		}
		return "", false, ctx.Err()
	default:
	}

	shortUrl, err = generator.GenerateShortUrl()
	if err != nil {
		return "", false, ErrGenerateShortUrl
	}

	if _, err = s.repo.GetByShort(ctx, shortUrl); err != nil {
		if !errors.Is(err, repository.ErrNotFoundShort) {
			return "", false, err
		}

		if err := s.repo.Create(ctx, shortUrl, originUrl); err != nil {
			switch {
			case errors.Is(err, repository.ErrDuplicateShort):
				return "", false, nil
			case errors.Is(err, repository.ErrDuplicateOrigin):
				existingShort, getErr := s.repo.GetByOrigin(ctx, originUrl)
				if getErr != nil {
					return "", false, getErr
				}
				return existingShort, true, nil
			default:
				return "", false, ErrPersistLink
			}
		}

		return shortUrl, true, nil
	}

	return "", false, nil
}
