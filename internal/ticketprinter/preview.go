package ticketprinter

import (
	"encoding/json"
	"fmt"
	"io"
)

const PreviewOutputPath = "preview.png"

func RunPreview(args []string, input io.Reader) error {
	if len(args) != 0 {
		return fmt.Errorf("preview does not accept arguments")
	}

	var payload WebhookPayload
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return fmt.Errorf("decode preview payload: %w", err)
	}

	label, err := payload.Label()
	if err != nil {
		return err
	}

	return WritePNG(PreviewOutputPath, label)
}
