package ticketprinter

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"

	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// Keep these files beside the renderer. Go embed cannot reach outside this package.
//
//go:embed assets/template.png
var templatePNG []byte

var (
	regularFont = mustParseFont(goregular.TTF)
	boldFont    = mustParseFont(gobold.TTF)
)

const (
	labelWidth  = 1181
	labelHeight = 566
)

const (
	qrX        = 19
	qrY        = 19
	qrSize     = 528
	textX      = 566
	textStartY = 220
	lineGap    = 10
	maxNameW   = 596
	nameSize   = 50
	bodySize   = 48
)

func RenderPNG(w io.Writer, label Label) error {
	img, err := renderLabel(label)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}

func WritePNG(path string, label Label) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output PNG directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create output PNG: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := RenderPNG(file, label); err != nil {
		return fmt.Errorf("write output PNG: %w", err)
	}
	return nil
}

func renderLabel(label Label) (image.Image, error) {
	base, err := png.Decode(bytes.NewReader(templatePNG))
	if err != nil {
		return nil, fmt.Errorf("decode template: %w", err)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, labelWidth, labelHeight))
	draw.Draw(canvas, canvas.Bounds(), base, image.Point{}, draw.Src)

	qr, err := qrcode.New(label.TicketURL, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("create QR code: %w", err)
	}
	qr.DisableBorder = true
	draw.Draw(canvas, image.Rect(qrX, qrY, qrX+qrSize, qrY+qrSize), qr.Image(qrSize), image.Point{}, draw.Over)

	regularFace, err := fontFace(regularFont, bodySize)
	if err != nil {
		return nil, err
	}
	defer closeFace(regularFace)

	boldFace, err := fittingBoldFace(label.RequesterName)
	if err != nil {
		return nil, err
	}
	defer closeFace(boldFace)

	y := textStartY
	drawText(canvas, textX, y, label.RequesterName, boldFace)
	y += nameSize + lineGap

	if label.Type != "" {
		drawText(canvas, textX, y, "Type: "+label.Type, regularFace)
		y += bodySize + lineGap
	}
	drawText(canvas, textX, y, "HelpDesk #: "+label.TicketNumber, regularFace)

	if label.CompNowTicketNo != "" {
		y += bodySize + lineGap
		drawText(canvas, textX, y, "CompNow #: "+label.CompNowTicketNo, regularFace)
	}

	drawTextBottomAligned(canvas, textX, qrY+qrSize, label.Date, regularFace)

	return canvas, nil
}

func fittingBoldFace(text string) (font.Face, error) {
	for size := float64(nameSize); size > 8; size-- {
		face, err := fontFace(boldFont, size)
		if err != nil {
			return nil, err
		}
		if font.MeasureString(face, text).Ceil() <= maxNameW {
			return face, nil
		}
		closeFace(face)
	}
	return fontFace(boldFont, 8)
}

func mustParseFont(fontBytes []byte) *sfnt.Font {
	parsed, err := opentype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
	return parsed
}

func fontFace(parsed *sfnt.Font, size float64) (font.Face, error) {
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("create font face: %w", err)
	}
	return face, nil
}

func closeFace(face font.Face) {
	if closer, ok := face.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}

func drawText(dst draw.Image, x, y int, text string, face font.Face) {
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.Black,
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(text)
}

func drawTextBottomAligned(dst draw.Image, x, bottomY int, text string, face font.Face) {
	metrics := face.Metrics()
	descent := metrics.Descent.Ceil()
	drawText(dst, x, bottomY-descent, text, face)
}
