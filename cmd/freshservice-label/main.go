package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/woodleighschool/freshservice-label/internal/ticketprinter"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(os.Args, logger); err != nil {
		logger.Error("freshservice-label failed", "err", err)
		os.Exit(1)
	}
}

func run(args []string, logger *slog.Logger) error {
	if len(args) != 1 {
		return errors.New("freshservice-label does not accept arguments")
	}

	cfg, err := ticketprinter.LoadConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	renderer, err := ticketprinter.NewRendererFromURL(ctx, cfg.LogoURL)
	if err != nil {
		return fmt.Errorf("initialize renderer: %w", err)
	}

	printer := ticketprinter.NewBrotherPrinter(cfg.PrinterAddr, cfg.PrintTimeout, logger)
	app := ticketprinter.NewServer(cfg, renderer, printer, logger)
	defer app.Close()

	router := chi.NewRouter()
	app.Routes(router)

	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "freshservice-label listening", "addr", cfg.ListenAddr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	}
}
