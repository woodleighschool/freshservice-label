package ticketprinter

import (
	"context"
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
	printer *brotherql.LabelPrinter
	timeout time.Duration
}

func NewBrotherPrinter(ctx context.Context, addr string, timeout time.Duration) (*BrotherPrinter, func(), error) {
	printer, err := brotherql.NewLabelPrinter(ctx, printerModel, printerBackend, addr)
	if err != nil {
		return nil, nil, err
	}

	closePrinter := func() {
		if err := printer.Close(); err != nil {
			slog.Error("close printer failed", "err", err)
		}
	}

	return &BrotherPrinter{printer: printer, timeout: timeout}, closePrinter, nil
}

func (p *BrotherPrinter) Print(ctx context.Context, label Label) error {
	img, err := renderLabel(label)
	if err != nil {
		return err
	}

	printCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	opts := brotherql.NewDefaultOptions(printerLabel)
	opts.Cut = true
	opts.Dither = false
	opts.Compress = false
	opts.Red = false
	opts.Rotate = printerRotate
	opts.Dpi600 = false
	opts.Hq = true
	opts.Threshold = printerThreshold

	return p.printer.Print(printCtx, []image.Image{img}, opts)
}
