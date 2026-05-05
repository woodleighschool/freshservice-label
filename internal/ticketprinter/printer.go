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

func NewBrotherPrinter(ctx context.Context, addr string, timeout time.Duration) (*BrotherPrinter, func(), error) {
	_ = ctx

	printer := &BrotherPrinter{
		addr:    addr,
		timeout: timeout,
		logger:  slog.Default(),
	}

	return printer, func() {}, nil
}

func (p *BrotherPrinter) Print(ctx context.Context, label Label) error {
	img, err := renderLabel(label)
	if err != nil {
		return err
	}

	printCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	printer, err := brotherql.NewLabelPrinter(printCtx, printerModel, printerBackend, p.addr)
	if err != nil {
		return fmt.Errorf("create printer: %w", err)
	}

	opts := printOptions()
	start := time.Now()
	p.logger.Info("printer write started", "ticket", label.TicketNumber, "addr", p.addr, "timeout", p.timeout)

	printErr := printer.Print(printCtx, []image.Image{img}, opts)
	if printErr == nil {
		p.logger.Info("printer write completed", "ticket", label.TicketNumber, "duration", time.Since(start))
	}

	closeErr := printer.Close()
	if closeErr != nil {
		p.logger.Error("printer connection close failed", "ticket", label.TicketNumber, "err", closeErr)
	} else {
		p.logger.Info("printer connection closed", "ticket", label.TicketNumber)
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
