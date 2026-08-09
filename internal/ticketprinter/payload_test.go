package ticketprinter

import (
	"encoding/json"
	"testing"
)

func TestLabelNormalizeOmitsRowsWithoutValues(t *testing.T) {
	var label Label
	if err := json.Unmarshal([]byte(`{
		"reference": "1234",
		"qr_url": "https://freshservice.example/a/tickets/1234",
		"title": "Example Person",
		"rows": [
			{"label": "Missing"},
			{"label": "Null", "value": null},
			{"label": "Empty", "value": ""},
			{"label": "Whitespace", "value": "   "},
			{"label": "Keep", "value": " value "}
		]
	}`), &label); err != nil {
		t.Fatalf("decode label: %v", err)
	}

	normalized, err := label.normalize()
	if err != nil {
		t.Fatalf("normalize label: %v", err)
	}
	if len(normalized.Rows) != 1 {
		t.Fatalf("normalized rows = %v, want one row", normalized.Rows)
	}
	if got, want := normalized.Rows[0], (Row{Label: "Keep", Value: "value"}); got != want {
		t.Fatalf("normalized row = %#v, want %#v", got, want)
	}
}

func TestLabelNormalizeRejectsMoreThanThreeRows(t *testing.T) {
	label := exampleLabel()
	label.Rows = append(label.Rows, Row{Label: "Extra", Value: "value"})

	_, err := label.normalize()
	if err == nil || err.Error() != "rows must contain at most 3 non-empty values" {
		t.Fatalf("normalize error = %v, want row limit error", err)
	}
}
