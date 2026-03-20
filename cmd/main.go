package main

import (
	"URLshortener/internal/handler"
	"URLshortener/internal/repository"
	"URLshortener/internal/repository/in_memory"
	"URLshortener/internal/service"
	"URLshortener/internal/tx"
	"URLshortener/pkg/database"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const ShutdownTimeout = 5 * time.Second

const defaultLRUCapacity = 10000
const defaultPort = "8080"

type Config struct {
	Storage     string
	LRUCapacity int
	Port        string
	DB          database.Config
}

func loadConfig() (Config, error) {
	storageFlag := flag.String("storage", "", "storage backend: memory or postgres")
	flag.Parse()

	cfg := Config{
		Storage: strings.ToLower(*storageFlag),
		Port:    strings.TrimSpace(os.Getenv("PORT")),
		DB: database.Config{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			Name:     os.Getenv("POSTGRES_DB"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
		},
	}
	if cfg.Storage == "" {
		return Config{}, errors.New("storage backend is required: use -storage=memory or -storage=postgres")
	}

	if cfg.Port == "" {
		cfg.Port = defaultPort
	} else {
		_, err := strconv.Atoi(cfg.Port)
		if err != nil {
			return Config{}, fmt.Errorf("invalid PORT: %q", cfg.Port)
		}
	}

	switch cfg.Storage {
	case "memory":
		rawCap := os.Getenv("IN_MEMORY_LRU_CAPACITY")
		if rawCap == "" {
			cfg.LRUCapacity = defaultLRUCapacity
			return cfg, nil
		}

		lruCap, err := strconv.Atoi(rawCap)
		if err != nil || lruCap < 0 {
			return Config{}, fmt.Errorf("invalid IN_MEMORY_LRU_CAPACITY: %q", rawCap)
		}
		cfg.LRUCapacity = lruCap
		return cfg, nil
	case "postgres":
		return cfg, nil
	default:
		return Config{}, errors.New("please choose storage mode between memory and postgres")
	}
}

func run(ctx context.Context, cfg Config) error {
	var serv *service.ShortenerService

	if cfg.Storage == "memory" {
		repo := in_memory.NewInMemoryRepository(cfg.LRUCapacity)
		txManager := tx.NewInMemoryTxManager()
		serv = service.NewShortenerService(txManager, repo)
	} else {
		dbpool, err := database.InitDb(ctx, cfg.DB)
		if err != nil {
			return fmt.Errorf("unable to connect to database: %w", err)
		}
		defer dbpool.Close()

		txManager := tx.NewPgxTxManager(dbpool)
		repo := repository.NewPostgresRepository(txManager)
		serv = service.NewShortenerService(txManager, repo)
	}

	handl := handler.NewShortenerHandler(serv)
	router := handler.NewRouter(handl)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Printf("url-shortener started on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()

		log.Println("Shutting down url-shortener service")

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen: %w", err)
	}
}

func main() {
	_ = godotenv.Load(".env")

	cfg, err := loadConfig()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
