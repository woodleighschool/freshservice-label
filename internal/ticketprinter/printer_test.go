package ticketprinter

import (
	"context"
	"errors"
	"io"
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

	printer, closePrinter, err := NewBrotherPrinter(context.Background(), listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatalf("new printer: %v", err)
	}
	t.Cleanup(closePrinter)

	label := Label{
		TicketURL:       "https://freshservice.example/a/tickets/1234",
		TicketNumber:    "1234",
		RequesterName:   "CODEX TEST",
		Type:            "Repair",
		CompNowTicketNo: "CN1234",
		Date:            "05 May 2026",
	}
	if err := printer.Print(context.Background(), label); err != nil {
		t.Fatalf("print: %v", err)
	}

	if err := <-readDone; err != nil {
		t.Fatalf("printer connection was not closed cleanly after print: %v", err)
	}
}
