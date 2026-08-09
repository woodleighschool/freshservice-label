package ticketprinter

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestRendererUsesLogoAreaForContentWhenLogoIsUnset(t *testing.T) {
	label := exampleLabel()

	withoutLogo := NewRenderer(nil)
	withoutLogoImage, err := withoutLogo.renderLabel(label)
	if err != nil {
		t.Fatalf("render without logo: %v", err)
	}
	if !hasInk(withoutLogoImage, withoutLogo.layout.logoBounds) {
		t.Fatal("logo area contains no text when logo is unset")
	}

	logo := image.NewRGBA(image.Rect(0, 0, 100, 20))
	draw.Draw(logo, logo.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	withLogo := NewRenderer(logo)
	withLogoImage, err := withLogo.renderLabel(label)
	if err != nil {
		t.Fatalf("render with logo: %v", err)
	}
	if hasInk(withLogoImage, withLogo.layout.logoBounds) {
		t.Fatal("text overlaps the logo area when a logo is configured")
	}
}

func TestRendererKeepsReferenceSlipDimensions(t *testing.T) {
	renderer := NewRenderer(nil)
	img, err := renderer.renderLabel(exampleLabel())
	if err != nil {
		t.Fatalf("render label: %v", err)
	}
	if got, want := img.Bounds().Size(), image.Pt(1181, 566); got != want {
		t.Fatalf("rendered size = %v, want %v", got, want)
	}
}

func exampleLabel() Label {
	return Label{
		Reference: "12345",
		QRURL:     "https://freshservice.example/a/tickets/12345",
		Title:     "Example Person",
		Rows: []Row{
			{Label: "Type", Value: "Repair"},
			{Label: "Ticket #", Value: "12345"},
			{Label: "Priority", Value: "Medium"},
		},
		Footer: "05 May 2026",
	}
}

func hasInk(img image.Image, bounds image.Rectangle) bool {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			red, green, blue, _ := img.At(x, y).RGBA()
			if red < 0xf000 || green < 0xf000 || blue < 0xf000 {
				return true
			}
		}
	}
	return false
}
