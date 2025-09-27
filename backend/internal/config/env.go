package config

import (
	"sync"

	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type AppEnv string

const (
	AppEnvDev  AppEnv = "dev"
	AppEnvProd AppEnv = "prod"
)

// EnvConfig holds all environment variables for the application
type EnvConfig struct {
	Port    string `env:"PORT" envDefault:"8080"`
	Debug   bool   `env:"DEBUG" envDefault:"false"`
	AppEnv  AppEnv `env:"APP_ENV" envDefault:"prod"`
	AppName string `env:"APP_NAME" envDefault:"stargo"`
}

var (
	appConfig *EnvConfig
	once      sync.Once
)

// loadConfig loads and validates all environment variables
func loadConfig() (*EnvConfig, error) {
	cfg := &EnvConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// InitConfig initializes the config only once
func InitConfig() (*EnvConfig, error) {
	var err error
	once.Do(func() {
		appConfig, err = loadConfig()
		zap.L().Info("Config loaded")
	})
	return appConfig, err
}

// Env returns the config. Panics if not initialized.
func Env() *EnvConfig {
	if appConfig == nil {
		zap.L().Panic("config not initialized — call InitConfig() first")
	}
	return appConfig
}
