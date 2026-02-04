package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/anti-brute-force/internal/config"
	"github.com/timutkin/anti-brute-force/internal/handler"
	"github.com/timutkin/anti-brute-force/internal/server"
	"github.com/timutkin/anti-brute-force/internal/service"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
	}
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Info().Err(err).Msg("failed to parse LOGGING_LEVEL, was set info level")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	blackListService := service.NewInMemoryListService()
	whiteListService := service.NewInMemoryListService()

	blackListHandler := handler.NewBlackListHandler(handler.NewListHandler(blackListService))
	whiteListHandler := handler.NewWhiteListHandler(handler.NewListHandler(whiteListService))
	bruteForceService := service.NewInMemoryBruteForceService(blackListService, whiteListService, cfg.Attempts)
	s := server.NewServer(blackListHandler, whiteListHandler, handler.NewBruteForceHandler(bruteForceService))

	address := cfg.Server.ListenAddress

	srv := &http.Server{
		Addr:              address,
		Handler:           s.GetHandler(),
		ReadHeaderTimeout: time.Second * 5,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msgf("listen %s", cfg.Server.ListenAddress)
		}
	}()

	log.Info().Msg("server starts ...")
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	<-ctx.Done()
	log.Info().Msg("shutdown server ...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("server exiting")
}
