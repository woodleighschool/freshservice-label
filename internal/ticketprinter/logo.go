package ticketprinter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	logoFetchTimeout = 10 * time.Second
	maxLogoBytes     = 5 << 20
	maxLogoPixels    = 16 << 20
	maxLogoRedirects = 3
)

// NewRendererFromURL creates a renderer and fetches the optional PNG logo once.
func NewRendererFromURL(ctx context.Context, rawLogoURL string) (*Renderer, error) {
	logo, err := fetchLogo(ctx, rawLogoURL)
	if err != nil {
		return nil, err
	}
	return NewRenderer(logo), nil
}

func fetchLogo(ctx context.Context, rawLogoURL string) (image.Image, error) {
	rawLogoURL = strings.TrimSpace(rawLogoURL)
	if rawLogoURL == "" {
		return nil, nil
	}

	parsed, err := url.Parse(rawLogoURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("LOGO_URL must be an absolute HTTP or HTTPS URL")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create logo request: %w", err)
	}

	client := &http.Client{
		Timeout: logoFetchTimeout,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= maxLogoRedirects {
				return errors.New("too many redirects")
			}
			return nil
		},
	}
	response, err := client.Do(request)
	if err != nil {
		var urlError *url.Error
		if errors.As(err, &urlError) {
			err = urlError.Err
		}
		return nil, fmt.Errorf("fetch logo: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch logo: unexpected HTTP status %d", response.StatusCode)
	}
	if response.ContentLength > maxLogoBytes {
		return nil, fmt.Errorf("fetch logo: PNG exceeds %d bytes", maxLogoBytes)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, maxLogoBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read logo: %w", err)
	}
	if len(data) > maxLogoBytes {
		return nil, fmt.Errorf("fetch logo: PNG exceeds %d bytes", maxLogoBytes)
	}

	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode logo PNG configuration: %w", err)
	}
	if config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > maxLogoPixels {
		return nil, fmt.Errorf("decode logo PNG: dimensions %dx%d are not supported", config.Width, config.Height)
	}

	logo, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode logo PNG: %w", err)
	}
	return logo, nil
}
