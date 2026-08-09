package ticketprinter

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestBrotherPrinterClosesNetworkConnectionAfterPrint(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
	})

	readDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			readDone <- err
			return
		}
		defer func() {
			_ = conn.Close()
		}()

		if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			readDone <- err
			return
		}

		buf := make([]byte, 32*1024)
		total := 0
		for {
			n, err := conn.Read(buf)
			total += n
			if errors.Is(err, io.EOF) {
				if total == 0 {
					readDone <- errors.New("printer received no bytes")
					return
				}
				readDone <- nil
				return
			}
			if err != nil {
				readDone <- err
				return
			}
		}
	}()

	printer := NewBrotherPrinter(listener.Addr().String(), 2*time.Second, slog.New(slog.DiscardHandler))

	label := Label{
		Reference: "1234",
		QRURL:     "https://freshservice.example/a/tickets/1234",
		Title:     "CODEX TEST",
		Rows: []Row{
			{Label: "Type", Value: "Repair"},
			{Label: "Ticket #", Value: "1234"},
		},
		Footer: "05 May 2026",
	}
	img, err := NewRenderer(nil).renderLabel(label)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if err := printer.Print(context.Background(), img, label.Reference); err != nil {
		t.Fatalf("print: %v", err)
	}

	if err := <-readDone; err != nil {
		t.Fatalf("printer connection was not closed cleanly after print: %v", err)
	}
}
