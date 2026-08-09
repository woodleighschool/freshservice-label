package ticketprinter

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestRoutesRegisterOnChiRouter(t *testing.T) {
	app := NewServer(Config{WebhookToken: "secret"}, NewRenderer(nil), nil, slog.New(slog.DiscardHandler))
	t.Cleanup(app.Close)

	router := chi.NewRouter()
	app.Routes(router)

	t.Run("healthz", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("GET /healthz status = %d, want %d", response.Code, http.StatusOK)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("POST /healthz status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestWebhookResponsesWaitForEachQueuedPrint(t *testing.T) {
	printer := newBlockingPrinter()
	app := NewServer(Config{WebhookToken: "secret", QueueDepth: 2}, NewRenderer(nil), printer, slog.New(slog.DiscardHandler))
	t.Cleanup(app.Close)

	router := chi.NewRouter()
	app.Routes(router)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	first := postWebhook(t, server.URL, "secret", webhookPayload("1001"))
	if got := printer.waitForStarted(t); got != "1001" {
		t.Fatalf("first started ticket = %q, want 1001", got)
	}
	assertNoResponse(t, first, "first webhook responded before first print completed")

	second := postWebhook(t, server.URL, "secret", webhookPayload("1002"))
	assertNoResponse(t, second, "second webhook responded before second print completed")

	printer.releasePrint()
	assertResponse(t, first, http.StatusOK)

	if got := printer.waitForStarted(t); got != "1002" {
		t.Fatalf("second started ticket = %q, want 1002", got)
	}
	assertNoResponse(t, second, "second webhook responded after first print completed, before second print completed")

	printer.releasePrint()
	assertResponse(t, second, http.StatusOK)

	if got, want := printer.printedTickets(), []string{"1001", "1002"}; !equalStrings(got, want) {
		t.Fatalf("printed tickets = %v, want %v", got, want)
	}
}

func TestWebhookRejectsContentThatDoesNotFitBeforePrinting(t *testing.T) {
	printer := &countingPrinter{}
	app := NewServer(Config{WebhookToken: "secret"}, NewRenderer(nil), printer, slog.New(slog.DiscardHandler))
	t.Cleanup(app.Close)

	label := webhookPayload("1001")
	label.Rows[0].Value = strings.Repeat("too wide ", 100)
	body, err := json.Marshal(label)
	if err != nil {
		t.Fatalf("marshal label: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer secret")
	response := httptest.NewRecorder()

	router := chi.NewRouter()
	app.Routes(router)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, response.Body.String())
	}
	if got := printer.prints.Load(); got != 0 {
		t.Fatalf("printer calls = %d, want 0", got)
	}
}

type countingPrinter struct {
	prints atomic.Int32
}

func (p *countingPrinter) Print(_ context.Context, _ image.Image, _ string) error {
	p.prints.Add(1)
	return nil
}

type blockingPrinter struct {
	started chan string
	release chan struct{}

	mu      sync.Mutex
	printed []string
}

func newBlockingPrinter() *blockingPrinter {
	return &blockingPrinter{
		started: make(chan string, 2),
		release: make(chan struct{}),
	}
}

func (p *blockingPrinter) Print(_ context.Context, _ image.Image, reference string) error {
	p.started <- reference
	<-p.release

	p.mu.Lock()
	defer p.mu.Unlock()
	p.printed = append(p.printed, reference)
	return nil
}

func (p *blockingPrinter) waitForStarted(t *testing.T) string {
	t.Helper()

	select {
	case ticket := <-p.started:
		return ticket
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for print to start")
		return ""
	}
}

func (p *blockingPrinter) releasePrint() {
	p.release <- struct{}{}
}

func (p *blockingPrinter) printedTickets() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	printed := make([]string, len(p.printed))
	copy(printed, p.printed)
	return printed
}

type webhookResult struct {
	status int
	body   string
	err    error
}

func postWebhook(t *testing.T, baseURL, token string, payload Label) <-chan webhookResult {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	result := make(chan webhookResult, 1)
	go func() {
		request, err := http.NewRequest(http.MethodPost, baseURL+"/webhook", bytes.NewReader(body))
		if err != nil {
			result <- webhookResult{err: err}
			return
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", "application/json")

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			result <- webhookResult{err: err}
			return
		}
		defer func() {
			_ = response.Body.Close()
		}()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			result <- webhookResult{err: err}
			return
		}

		result <- webhookResult{status: response.StatusCode, body: string(responseBody)}
	}()
	return result
}

func webhookPayload(ticket string) Label {
	return Label{
		Reference: ticket,
		QRURL:     "https://freshservice.example/a/tickets/" + ticket,
		Title:     "Queue Test",
		Rows: []Row{
			{Label: "Type", Value: "Repair"},
			{Label: "Ticket #", Value: ticket},
			{Label: "Vendor #", Value: "VN" + ticket},
		},
		Footer: "05 May 2026",
	}
}

func assertNoResponse(t *testing.T, result <-chan webhookResult, message string) {
	t.Helper()

	select {
	case got := <-result:
		if got.err != nil {
			t.Fatalf("%s: request failed early: %v", message, got.err)
		}
		t.Fatalf("%s: got HTTP %d", message, got.status)
	case <-time.After(100 * time.Millisecond):
		return
	}
}

func assertResponse(t *testing.T, result <-chan webhookResult, want int) {
	t.Helper()

	select {
	case got := <-result:
		if got.err != nil {
			t.Fatalf("request failed: %v", got.err)
		}
		if got.status != want {
			t.Fatalf("status = %d, want %d; body = %q", got.status, want, got.body)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for HTTP %d", want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
