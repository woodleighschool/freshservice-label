# freshservice-label

Small webhook service for printing our Freshservice labels on the Brother QL-820NWB.

Freshservice sends this service a compact JSON payload from an automation. The service queues the webhook in memory, renders the label, prints directly to the Ethernet-connected printer over TCP/9100, then responds to that webhook after the label has printed.

![Example Label](example.png)

## Run

```sh
WEBHOOK_TOKEN=secret \
PRINTER_ADDR=172.19.10.13 \
go run ./cmd/freshservice-label
```

The default listen address is `:8080`.

## Configuration

| Variable | Required | Default | Notes |
| --- | --- | --- | --- |
| `WEBHOOK_TOKEN` | yes | | Bearer token expected from Freshservice. |
| `PRINTER_ADDR` | yes | | Printer address, such as `172.19.10.13` or `tcp://172.19.10.13:9100`. |
| `LISTEN_ADDR` | no | `:8080` | HTTP listen address. |
| `QUEUE_DEPTH` | no | `10` | In-memory webhook queue depth. Jobs vanish on restart. |
| `PRINT_TIMEOUT` | no | `30s` | Per-label print timeout. |

## Webhook Payload

Freshservice can shape the payload, so this app only knows the fields it needs:

```json
{
  "ticket_url": "https://freshservice.example/a/tickets/12345",
  "requester_name": "Example Person",
  "subject": "REPAIR - laptop issue",
  "created_at": "2026-05-05T02:35:00Z",
  "compnow_ticket_no": "CN123456"
}
```

`compnow_ticket_no` is optional. The QR code uses `ticket_url`; the printed HelpDesk number is the final path segment of that URL.

Example:

```sh
curl -X POST http://127.0.0.1:8080/webhook \
  -H 'Authorization: Bearer secret' \
  -H 'Content-Type: application/json' \
  -d '{
    "ticket_url": "https://freshservice.example/a/tickets/12345",
    "requester_name": "Example Person",
    "subject": "REPAIR - laptop issue",
    "created_at": "2026-05-05T02:35:00Z",
    "compnow_ticket_no": "CN123456"
  }'
```

## Development

```sh
make check
```

To preview the label without starting the service or using stock:

```sh
make preview
```

This reads `preview.json` and writes `preview.png` in the repo root. Preview mode only renders the PNG; it does not read service env vars or contact the printer.

The label layout lives in commented constants in `internal/ticketprinter/render.go`. Adjust those before reaching for a larger templating system.
