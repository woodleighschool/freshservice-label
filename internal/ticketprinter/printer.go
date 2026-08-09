package ticketprinter

import (
	"context"
	"errors"
	"fmt"
	"image"
	"log/slog"
	"time"

	brotherql "github.com/suapapa/go_brother-ql"
)

const (
	printerModel     = "QL-820NWB"
	printerBackend   = "network"
	printerLabel     = "62"
	printerRotate    = "0"
	printerThreshold = 70.0
)

type BrotherPrinter struct {
	addr    string
	timeout time.Duration
	logger  *slog.Logger
}

func NewBrotherPrinter(addr string, timeout time.Duration, logger *slog.Logger) *BrotherPrinter {
	return &BrotherPrinter{
		addr:    addr,
		timeout: timeout,
		logger:  logger,
	}
}

func (p *BrotherPrinter) Print(ctx context.Context, img image.Image, reference string) error {
	printCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	printer, err := brotherql.NewLabelPrinter(printCtx, printerModel, printerBackend, p.addr)
	if err != nil {
		return fmt.Errorf("create printer: %w", err)
	}

	opts := printOptions()
	start := time.Now()
	p.logger.InfoContext(ctx, "printer write started", "reference", reference, "addr", p.addr, "timeout", p.timeout)

	printErr := printer.Print(printCtx, []image.Image{img}, opts)
	if printErr == nil {
		p.logger.InfoContext(ctx, "printer write completed", "reference", reference, "duration", time.Since(start))
	}

	closeErr := printer.Close()
	if closeErr != nil {
		p.logger.ErrorContext(ctx, "printer connection close failed", "reference", reference, "err", closeErr)
	} else {
		p.logger.InfoContext(ctx, "printer connection closed", "reference", reference)
	}

	if printErr != nil || closeErr != nil {
		return errors.Join(printErr, closeErr)
	}
	return nil
}

func printOptions() brotherql.PrintOptions {
	opts := brotherql.NewDefaultOptions(printerLabel)
	opts.Cut = true
	opts.Dither = false
	opts.Compress = false
	opts.Red = false
	opts.Rotate = printerRotate
	opts.Dpi600 = false
	opts.Hq = true
	opts.Threshold = printerThreshold

	return opts
}
