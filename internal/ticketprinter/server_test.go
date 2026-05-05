package ticketprinter

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

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
