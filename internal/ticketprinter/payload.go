package ticketprinter

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type WebhookPayload struct {
	TicketURL       string `json:"ticket_url"`
	RequesterName   string `json:"requester_name"`
	Subject         string `json:"subject"`
	CreatedAt       string `json:"created_at"`
	CompNowTicketNo string `json:"compnow_ticket_no"`
}

type Label struct {
	TicketURL       string
	TicketNumber    string
	RequesterName   string
	Type            string
	CompNowTicketNo string
	Date            string
}

func (p WebhookPayload) Label() (Label, error) {
	p.TicketURL = strings.TrimSpace(p.TicketURL)
	p.RequesterName = strings.TrimSpace(p.RequesterName)
	p.Subject = strings.TrimSpace(p.Subject)
	p.CreatedAt = strings.TrimSpace(p.CreatedAt)
	p.CompNowTicketNo = strings.TrimSpace(p.CompNowTicketNo)

	switch {
	case p.TicketURL == "":
		return Label{}, errors.New("ticket_url is required")
	case p.RequesterName == "":
		return Label{}, errors.New("requester_name is required")
	case p.Subject == "":
		return Label{}, errors.New("subject is required")
	case p.CreatedAt == "":
		return Label{}, errors.New("created_at is required")
	}

	ticketNumber, err := ticketNumberFromURL(p.TicketURL)
	if err != nil {
		return Label{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		return Label{}, fmt.Errorf("created_at must be RFC3339: %w", err)
	}

	return Label{
		TicketURL:       p.TicketURL,
		TicketNumber:    ticketNumber,
		RequesterName:   p.RequesterName,
		Type:            ticketType(p.Subject),
		CompNowTicketNo: p.CompNowTicketNo,
		Date:            createdAt.Local().Format("02 Jan 2006"),
	}, nil
}

func ticketNumberFromURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("ticket_url must be an absolute URL")
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) == 0 || parts[len(parts)-1] == "" {
		return "", errors.New("ticket_url must end with a ticket identifier")
	}
	return parts[len(parts)-1], nil
}

func ticketType(subject string) string {
	subject = strings.ToLower(subject)

	switch {
	case strings.Contains(subject, "repair"):
		return "Repair"
	case strings.Contains(subject, "year 12 wipe"):
		return "Year 12 Wipe"
	case strings.Contains(subject, "machine returned"):
		return "Return"
	default:
		return ""
	}
}
