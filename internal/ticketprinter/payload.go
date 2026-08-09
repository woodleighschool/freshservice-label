package ticketprinter

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const maxRows = 3

// Row is one optional label-value line on a label.
type Row struct {
	Label string `json:"label,omitempty"`
	Value string `json:"value"`
}

// Label contains the display-ready content supplied by the webhook.
type Label struct {
	Reference string `json:"reference"`
	QRURL     string `json:"qr_url"`
	Title     string `json:"title"`
	Rows      []Row  `json:"rows,omitempty"`
	Footer    string `json:"footer,omitempty"`
}

func (label Label) normalize() (Label, error) {
	label.Reference = strings.TrimSpace(label.Reference)
	label.QRURL = strings.TrimSpace(label.QRURL)
	label.Title = strings.TrimSpace(label.Title)
	label.Footer = strings.TrimSpace(label.Footer)

	switch {
	case label.Reference == "":
		return Label{}, errors.New("reference is required")
	case label.QRURL == "":
		return Label{}, errors.New("qr_url is required")
	case label.Title == "":
		return Label{}, errors.New("title is required")
	}

	parsed, err := url.Parse(label.QRURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return Label{}, errors.New("qr_url must be an absolute URL")
	}

	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "reference", value: label.Reference},
		{name: "title", value: label.Title},
		{name: "footer", value: label.Footer},
	} {
		if strings.ContainsAny(field.value, "\r\n") {
			return Label{}, fmt.Errorf("%s must be a single line", field.name)
		}
	}

	rows := make([]Row, 0, len(label.Rows))
	for i, row := range label.Rows {
		row.Label = strings.TrimSpace(row.Label)
		row.Value = strings.TrimSpace(row.Value)
		if row.Value == "" {
			continue
		}
		if strings.ContainsAny(row.Label, "\r\n") || strings.ContainsAny(row.Value, "\r\n") {
			return Label{}, fmt.Errorf("row %d must be a single line", i+1)
		}
		rows = append(rows, row)
	}
	label.Rows = rows
	if len(label.Rows) > maxRows {
		return Label{}, fmt.Errorf("rows must contain at most %d non-empty values", maxRows)
	}

	return label, nil
}

func (row Row) text() string {
	if row.Label == "" {
		return row.Value
	}
	return row.Label + ": " + row.Value
}
