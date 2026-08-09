package ticketprinter

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"image"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type Printer interface {
	Print(context.Context, image.Image, string) error
}

type Server struct {
	token    string
	logger   *slog.Logger
	renderer *Renderer
	printer  Printer
	jobs     chan printJob

	stop      chan struct{}
	stopped   chan struct{}
	closeOnce sync.Once
}

type printJob struct {
	reference string
	image     image.Image
	queuedAt  time.Time
	result    chan error
}

const maxWebhookBodyBytes = 64 << 10

func NewServer(cfg Config, renderer *Renderer, printer Printer, logger *slog.Logger) *Server {
	if cfg.QueueDepth < 1 {
		cfg.QueueDepth = 1
	}

	server := &Server{
		token:    cfg.WebhookToken,
		logger:   logger,
		renderer: renderer,
		printer:  printer,
		jobs:     make(chan printJob, cfg.QueueDepth),
		stop:     make(chan struct{}),
		stopped:  make(chan struct{}),
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
			ctx := context.Background()
			start := time.Now()
			s.logger.InfoContext(ctx, "print started", "reference", job.reference, "wait", start.Sub(job.queuedAt))

			err := s.printer.Print(ctx, job.image, job.reference)
			if err != nil {
				s.logger.ErrorContext(ctx, "print failed", "reference", job.reference, "duration", time.Since(start), "err", err)
			} else {
				s.logger.InfoContext(ctx, "print completed", "reference", job.reference, "duration", time.Since(start))
			}
			job.result <- err
		}
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(r.Context(), w, http.StatusOK, "ok", "")
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if !validBearer(r.Header.Get("Authorization"), s.token) {
		s.writeJSON(r.Context(), w, http.StatusUnauthorized, "failed", "unauthorized")
		return
	}

	var label Label
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&label); err != nil {
		s.writeJSON(r.Context(), w, http.StatusBadRequest, "failed", "invalid JSON")
		return
	}

	label, err := label.normalize()
	if err != nil {
		s.writeJSON(r.Context(), w, http.StatusBadRequest, "failed", err.Error())
		return
	}
	img, err := s.renderer.render(label)
	if err != nil {
		s.writeJSON(r.Context(), w, http.StatusBadRequest, "failed", err.Error())
		return
	}

	job := printJob{reference: label.Reference, image: img, queuedAt: time.Now(), result: make(chan error, 1)}
	select {
	case s.jobs <- job:
		s.logger.InfoContext(r.Context(), "print queued", "reference", label.Reference, "queue_depth", len(s.jobs), "queue_capacity", cap(s.jobs))
	default:
		s.logger.WarnContext(r.Context(), "print queue full", "reference", label.Reference, "queue_capacity", cap(s.jobs))
		s.writeJSON(r.Context(), w, http.StatusServiceUnavailable, "failed", "print queue is full")
		return
	}

	select {
	case err := <-job.result:
		if err != nil {
			s.writeJSON(r.Context(), w, http.StatusInternalServerError, "failed", err.Error())
			return
		}
		s.writeJSON(r.Context(), w, http.StatusOK, "success", "")
	case <-r.Context().Done():
		s.logger.InfoContext(r.Context(), "request cancelled", "reference", label.Reference)
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

func (s *Server) writeJSON(ctx context.Context, w http.ResponseWriter, code int, status, reason string) {
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
		s.logger.ErrorContext(ctx, "write response failed", "err", err)
	}
}
