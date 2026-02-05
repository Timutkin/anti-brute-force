package config

import (
	"context"

	"github.com/heetch/confita"
)

type AntiBruteForceConfig struct {
	Attempts AttemptsConfig
	Server   Server
	Logger   Logger
}

type AttemptsConfig struct {
	IPMaxAttempts       int `config:"IP_MAX_ATTEMPTS"`
	PasswordMaxAttempts int `config:"PASSWORD_MAX_ATTEMPTS"`
	LoginMaxAttempts    int `config:"LOGIN_MAX_ATTEMPTS"`
}

type Server struct {
	ListenAddress string `config:"LISTEN_ADDRESS"`
}

type Logger struct {
	Level string `config:"LOGGING_LEVEL"`
}

func NewConfig() (AntiBruteForceConfig, error) {
	cfg := AntiBruteForceConfig{
		Server: Server{ListenAddress: "localhost:8080"},
		Attempts: AttemptsConfig{
			IPMaxAttempts:       1000,
			PasswordMaxAttempts: 100,
			LoginMaxAttempts:    10,
		},
		Logger: Logger{Level: "INFO"},
	}
	loader := confita.NewLoader()
	err := loader.Load(context.Background(), &cfg)
	return cfg, err
}
