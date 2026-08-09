package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"

	"github.com/woodleighschool/freshservice-label/internal/ticketprinter"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("preview", flag.ContinueOnError)
	output := flags.String("output", "preview.png", "output PNG path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("preview does not accept positional arguments")
	}

	renderer := ticketprinter.NewRenderer(placeholderLogo())
	if logoURL := strings.TrimSpace(os.Getenv("LOGO_URL")); logoURL != "" {
		brandedRenderer, err := ticketprinter.NewRendererFromURL(context.Background(), logoURL)
		if err != nil {
			return err
		}
		renderer = brandedRenderer
	}

	label := ticketprinter.Label{
		Reference: "12345",
		QRURL:     "https://freshservice.example/a/tickets/12345",
		Title:     "Example Person",
		Rows: []ticketprinter.Row{
			{Label: "Type", Value: "Repair"},
			{Label: "Ticket #", Value: "12345"},
			{Label: "Priority", Value: "Medium"},
		},
		Footer: "05 May 2026",
	}

	if err := renderer.WritePNG(*output, label); err != nil {
		return err
	}
	_, err := fmt.Fprintln(os.Stdout, *output)
	return err
}

func placeholderLogo() image.Image {
	logo := image.NewRGBA(image.Rect(0, 0, 80, 20))
	draw.Draw(logo, logo.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	(&font.Drawer{
		Dst:  logo,
		Src:  image.Black,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(1, 16),
	}).DrawString("<Logo Here>")
	return logo
}
