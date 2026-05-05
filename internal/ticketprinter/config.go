package ticketprinter

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr   string
	WebhookToken string
	PrinterAddr  string
	QueueDepth   int
	PrintTimeout time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ListenAddr:   env("LISTEN_ADDR", ":8080"),
		WebhookToken: os.Getenv("WEBHOOK_TOKEN"),
		PrinterAddr:  os.Getenv("PRINTER_ADDR"),
		QueueDepth:   10,
		PrintTimeout: 30 * time.Second,
	}

	if cfg.WebhookToken == "" {
		return Config{}, errors.New("WEBHOOK_TOKEN is required")
	}
	if cfg.PrinterAddr == "" {
		return Config{}, errors.New("PRINTER_ADDR is required")
	}

	if v := os.Getenv("QUEUE_DEPTH"); v != "" {
		queueDepth, err := strconv.Atoi(v)
		if err != nil || queueDepth < 1 {
			return Config{}, fmt.Errorf("QUEUE_DEPTH must be a positive integer")
		}
		cfg.QueueDepth = queueDepth
	}

	if v := os.Getenv("PRINT_TIMEOUT"); v != "" {
		printTimeout, err := time.ParseDuration(v)
		if err != nil || printTimeout <= 0 {
			return Config{}, fmt.Errorf("PRINT_TIMEOUT must be a positive duration")
		}
		cfg.PrintTimeout = printTimeout
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
