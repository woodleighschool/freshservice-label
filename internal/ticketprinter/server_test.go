package ticketprinter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestRoutesRegisterOnChiRouter(t *testing.T) {
	app := NewServer(Config{WebhookToken: "secret"}, nil, slog.New(slog.DiscardHandler))
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
	app := NewServer(Config{WebhookToken: "secret", QueueDepth: 2}, printer, slog.New(slog.DiscardHandler))
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

func (p *blockingPrinter) Print(_ context.Context, label Label) error {
	p.started <- label.TicketNumber
	<-p.release

	p.mu.Lock()
	defer p.mu.Unlock()
	p.printed = append(p.printed, label.TicketNumber)
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

func postWebhook(t *testing.T, baseURL, token string, payload WebhookPayload) <-chan webhookResult {
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

func webhookPayload(ticket string) WebhookPayload {
	return WebhookPayload{
		TicketURL:       "https://freshservice.example/a/tickets/" + ticket,
		RequesterName:   "Queue Test",
		Subject:         "REPAIR - queue test",
		CreatedAt:       "2026-05-05T02:35:00Z",
		CompNowTicketNo: "CN" + ticket,
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
