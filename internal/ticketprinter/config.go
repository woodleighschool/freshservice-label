package ticketprinter

import (
	"errors"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains the service's environment-derived configuration.
type Config struct {
	ListenAddr   string        `env:"LISTEN_ADDR"   envDefault:":8080"`
	WebhookToken string        `env:"WEBHOOK_TOKEN,required,notEmpty"`
	PrinterAddr  string        `env:"PRINTER_ADDR,required,notEmpty"`
	LogoURL      string        `env:"LOGO_URL"`
	QueueDepth   int           `env:"QUEUE_DEPTH"   envDefault:"10"`
	PrintTimeout time.Duration `env:"PRINT_TIMEOUT" envDefault:"30s"`
}

// LoadConfig parses and validates configuration from the environment.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	if cfg.QueueDepth < 1 {
		return Config{}, errors.New("QUEUE_DEPTH must be a positive integer")
	}
	if cfg.PrintTimeout <= 0 {
		return Config{}, errors.New("PRINT_TIMEOUT must be a positive duration")
	}

	return cfg, nil
}
