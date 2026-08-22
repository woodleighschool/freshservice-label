package ticketprinter

import (
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("WEBHOOK_TOKEN", "secret")
	t.Setenv("PRINTER_ADDR", "tcp://printer.example:9100")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q, want :8080", cfg.ListenAddr)
	}
	if cfg.WebhookToken != "secret" {
		t.Errorf("WebhookToken = %q, want secret", cfg.WebhookToken)
	}
	if cfg.PrinterAddr != "tcp://printer.example:9100" {
		t.Errorf("PrinterAddr = %q, want tcp://printer.example:9100", cfg.PrinterAddr)
	}
	if cfg.QueueDepth != 10 {
		t.Errorf("QueueDepth = %d, want 10", cfg.QueueDepth)
	}
	if cfg.PrintTimeout != 30*time.Second {
		t.Errorf("PrintTimeout = %v, want 30s", cfg.PrintTimeout)
	}
}

func TestLoadConfigRejectsInvalidEnvironment(t *testing.T) {
	tests := map[string]struct {
		key   string
		value string
	}{
		"empty webhook token": {key: "WEBHOOK_TOKEN", value: ""},
		"empty printer address": {
			key:   "PRINTER_ADDR",
			value: "",
		},
		"non-numeric queue depth": {key: "QUEUE_DEPTH", value: "many"},
		"zero queue depth":        {key: "QUEUE_DEPTH", value: "0"},
		"invalid print timeout":   {key: "PRINT_TIMEOUT", value: "soon"},
		"zero print timeout":      {key: "PRINT_TIMEOUT", value: "0s"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("WEBHOOK_TOKEN", "secret")
			t.Setenv("PRINTER_ADDR", "tcp://printer.example:9100")
			t.Setenv(test.key, test.value)

			if _, err := LoadConfig(); err == nil {
				t.Fatal("LoadConfig() returned nil error")
			}
		})
	}
}
