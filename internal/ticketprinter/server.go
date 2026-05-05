package ticketprinter

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type Printer interface {
	Print(context.Context, Label) error
}

type Server struct {
	token   string
	logger  *slog.Logger
	printer Printer
	jobs    chan printJob

	stop      chan struct{}
	stopped   chan struct{}
	closeOnce sync.Once
}

type printJob struct {
	label    Label
	queuedAt time.Time
	result   chan error
}

func NewServer(cfg Config, printer Printer, logger *slog.Logger) *Server {
	if cfg.QueueDepth < 1 {
		cfg.QueueDepth = 1
	}
	if logger == nil {
		logger = slog.Default()
	}

	server := &Server{
		token:   cfg.WebhookToken,
		logger:  logger,
		printer: printer,
		jobs:    make(chan printJob, cfg.QueueDepth),
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	go server.worker()
	return server
}

func (s *Server) Routes(router chi.Router) {
	router.Get("/healthz", s.handleHealthz)
	router.Post("/webhook", s.handleWebhook)
}

func (s *Server) Close() {
	s.closeOnce.Do(func() {
		close(s.stop)
		<-s.stopped
	})
}

func (s *Server) worker() {
	defer close(s.stopped)

	for {
		select {
		case <-s.stop:
			return
		case job := <-s.jobs:
			start := time.Now()
			s.logger.Info("print started", "ticket", job.label.TicketNumber, "wait", start.Sub(job.queuedAt))

			err := s.printer.Print(context.Background(), job.label)
			if err != nil {
				s.logger.Error("print failed", "ticket", job.label.TicketNumber, "duration", time.Since(start), "err", err)
			} else {
				s.logger.Info("print completed", "ticket", job.label.TicketNumber, "duration", time.Since(start))
			}
			job.result <- err
		}
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, "ok", "")
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if !validBearer(r.Header.Get("Authorization"), s.token) {
		writeJSON(w, http.StatusUnauthorized, "failed", "unauthorized")
		return
	}

	var payload WebhookPayload
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, "failed", "invalid JSON")
		return
	}

	label, err := payload.Label()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, "failed", err.Error())
		return
	}

	job := printJob{label: label, queuedAt: time.Now(), result: make(chan error, 1)}
	select {
	case s.jobs <- job:
		s.logger.Info("print queued", "ticket", label.TicketNumber, "queue_depth", len(s.jobs), "queue_capacity", cap(s.jobs))
	default:
		s.logger.Warn("print queue full", "ticket", label.TicketNumber, "queue_capacity", cap(s.jobs))
		writeJSON(w, http.StatusServiceUnavailable, "failed", "print queue is full")
		return
	}

	select {
	case err := <-job.result:
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, "failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, "success", "")
	case <-r.Context().Done():
		s.logger.Info("request cancelled", "ticket", label.TicketNumber)
	}
}

func validBearer(header, token string) bool {
	const prefix = "Bearer "
	got, ok := strings.CutPrefix(header, prefix)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(got)), []byte(token)) == 1
}

func writeJSON(w http.ResponseWriter, code int, status, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := struct {
		Status string `json:"status"`
		Reason string `json:"reason,omitempty"`
	}{
		Status: status,
		Reason: reason,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("write response failed", "err", err)
	}
}
