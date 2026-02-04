package main

import (
	"context"
	"os/signal"
	"syscall"

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
	s := server.NewServer(blackListHandler, whiteListHandler, handler.NewBruteForceHandler(bruteForceService), cfg.Server)

	go func() {
		err := s.Start()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	log.Info().Msg("server starts ...")
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	<-ctx.Done()
}
