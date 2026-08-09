package ticketprinter

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"

	qrcode "github.com/skip2/go-qrcode"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var (
	regularFont = mustParseFont(goregular.TTF)
	boldFont    = mustParseFont(gobold.TTF)
)

const referenceSlipHeight = 566

type slipLayout struct {
	bounds            image.Rectangle
	qrBounds          image.Rectangle
	contentBounds     image.Rectangle
	logoBounds        image.Rectangle
	titleWithLogoY    int
	titleWithoutLogoY int
	titleSize         float64
	rowSize           float64
	titleAdvance      int
	rowAdvance        int
	lineGap           int
}

func newSlipLayout(height int) slipLayout {
	const (
		referenceWidth          = 1181
		referenceMargin         = 19
		referenceLogoBottom     = 149
		referenceLogoRight      = 1163
		referenceTitleWithLogoY = 220
		referenceTitleSize      = 50
		referenceRowSize        = 48
		referenceLineGap        = 10
	)

	scale := float64(height) / referenceSlipHeight
	px := func(value int) int {
		return int(math.Round(float64(value) * scale))
	}

	width := px(referenceWidth)
	margin := px(referenceMargin)
	qrSize := height - 2*margin
	contentX := margin + qrSize + margin
	contentBounds := image.Rect(contentX, margin, width-margin, height-margin)
	titleSize := float64(referenceTitleSize) * scale
	rowSize := float64(referenceRowSize) * scale
	lineGap := px(referenceLineGap)

	return slipLayout{
		bounds:            image.Rect(0, 0, width, height),
		qrBounds:          image.Rect(margin, margin, margin+qrSize, margin+qrSize),
		contentBounds:     contentBounds,
		logoBounds:        image.Rect(contentBounds.Min.X, margin, px(referenceLogoRight), px(referenceLogoBottom)),
		titleWithLogoY:    px(referenceTitleWithLogoY),
		titleWithoutLogoY: margin + int(math.Round(titleSize)),
		titleSize:         titleSize,
		rowSize:           rowSize,
		titleAdvance:      int(math.Round(titleSize)) + lineGap,
		rowAdvance:        int(math.Round(rowSize)) + lineGap,
		lineGap:           lineGap,
	}
}

// Renderer composes labels using the fixed landscape slip layout.
type Renderer struct {
	layout slipLayout
	logo   image.Image
}

// NewRenderer creates a renderer with an optional logo image.
func NewRenderer(logo image.Image) *Renderer {
	return &Renderer{
		layout: newSlipLayout(referenceSlipHeight),
		logo:   logo,
	}
}

// RenderPNG writes label as a PNG using the fixed slip layout.
func (r *Renderer) RenderPNG(w io.Writer, label Label) error {
	img, err := r.renderLabel(label)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}

// WritePNG renders label to path, creating its parent directory when needed.
func (r *Renderer) WritePNG(path string, label Label) error {
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

	if err := r.RenderPNG(file, label); err != nil {
		return fmt.Errorf("write output PNG: %w", err)
	}
	return nil
}

func (r *Renderer) renderLabel(label Label) (image.Image, error) {
	label, err := label.normalize()
	if err != nil {
		return nil, err
	}
	return r.render(label)
}

func (r *Renderer) render(label Label) (image.Image, error) {
	canvas := image.NewRGBA(r.layout.bounds)
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	qr, err := qrcode.New(label.QRURL, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("create QR code: %w", err)
	}
	qr.DisableBorder = true
	draw.Draw(canvas, r.layout.qrBounds, qr.Image(r.layout.qrBounds.Dx()), image.Point{}, draw.Over)

	if r.logo != nil {
		drawContained(canvas, r.layout.logoBounds, r.logo)
	}

	regularFace, err := fontFace(regularFont, r.layout.rowSize)
	if err != nil {
		return nil, err
	}
	defer closeFace(regularFace)

	titleFace, err := fittingFace(boldFont, label.Title, r.layout.titleSize, r.layout.contentBounds.Dx())
	if err != nil {
		return nil, err
	}
	defer closeFace(titleFace)

	for i, row := range label.Rows {
		if font.MeasureString(regularFace, row.text()).Ceil() > r.layout.contentBounds.Dx() {
			return nil, fmt.Errorf("row %d does not fit the label", i+1)
		}
	}
	if font.MeasureString(regularFace, label.Footer).Ceil() > r.layout.contentBounds.Dx() {
		return nil, errors.New("footer does not fit the label")
	}

	titleY := r.layout.titleWithoutLogoY
	if r.logo != nil {
		titleY = r.layout.titleWithLogoY
	}

	flowBottom := titleY + titleFace.Metrics().Descent.Ceil()
	if len(label.Rows) != 0 {
		lastRowY := titleY + r.layout.titleAdvance + (len(label.Rows)-1)*r.layout.rowAdvance
		flowBottom = lastRowY + regularFace.Metrics().Descent.Ceil()
	}

	flowLimit := r.layout.contentBounds.Max.Y
	footerY := 0
	if label.Footer != "" {
		metrics := regularFace.Metrics()
		footerY = r.layout.contentBounds.Max.Y - metrics.Descent.Ceil()
		flowLimit = footerY - metrics.Ascent.Ceil() - r.layout.lineGap
	}
	if flowBottom > flowLimit {
		return nil, errors.New("label content does not fit the slip")
	}

	drawText(canvas, r.layout.contentBounds.Min.X, titleY, label.Title, titleFace)
	rowY := titleY + r.layout.titleAdvance
	for _, row := range label.Rows {
		drawText(canvas, r.layout.contentBounds.Min.X, rowY, row.text(), regularFace)
		rowY += r.layout.rowAdvance
	}
	if label.Footer != "" {
		drawText(canvas, r.layout.contentBounds.Min.X, footerY, label.Footer, regularFace)
	}

	return canvas, nil
}

func fittingFace(parsed *sfnt.Font, text string, maxSize float64, maxWidth int) (font.Face, error) {
	minimumSize := maxSize * 8 / 50
	for size := maxSize; size >= minimumSize; size-- {
		face, err := fontFace(parsed, size)
		if err != nil {
			return nil, err
		}
		if font.MeasureString(face, text).Ceil() <= maxWidth {
			return face, nil
		}
		closeFace(face)
	}
	return nil, errors.New("title does not fit the label")
}

func drawContained(dst draw.Image, bounds image.Rectangle, src image.Image) {
	sourceBounds := src.Bounds()
	scale := math.Min(
		float64(bounds.Dx())/float64(sourceBounds.Dx()),
		float64(bounds.Dy())/float64(sourceBounds.Dy()),
	)
	width := max(1, int(math.Round(float64(sourceBounds.Dx())*scale)))
	height := max(1, int(math.Round(float64(sourceBounds.Dy())*scale)))
	x := bounds.Min.X + (bounds.Dx()-width)/2
	y := bounds.Min.Y + (bounds.Dy()-height)/2
	target := image.Rect(x, y, x+width, y+height)
	xdraw.CatmullRom.Scale(dst, target, src, sourceBounds, draw.Over, nil)
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
