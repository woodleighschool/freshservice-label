# freshservice-label

Webhook service that prints Freshservice labels on a Brother QL-820NWB using the fixed landscape layout for 62 mm continuous stock.

![Example label](example.png)

## Run

```sh
WEBHOOK_TOKEN=secret \
PRINTER_ADDR=172.19.10.13 \
go run ./cmd/freshservice-label
```

`LOGO_URL` is optional. When set, the PNG is fetched during startup and held in memory.

| Variable        | Required | Default |
| --------------- | -------- | ------- |
| `WEBHOOK_TOKEN` | yes      |         |
| `PRINTER_ADDR`  | yes      |         |
| `LOGO_URL`      | no       | no logo |
| `LISTEN_ADDR`   | no       | `:8080` |
| `QUEUE_DEPTH`   | no       | `10`    |
| `PRINT_TIMEOUT` | no       | `30s`   |

## Webhook

Use an Advanced JSON webhook in Freshservice:

```json
{
    "reference": "{{ticket.id_numeric}}",
    "qr_url": "{{ticket.url}}",
    "title": "{{ticket.requester.name}}",
    "rows": [
        { "label": "Type", "value": "{{ticket.ticket_type}}" },
        { "label": "Ticket #", "value": "{{ticket.id_numeric}}" },
        { "label": "Priority", "value": "{{ticket.priority}}" }
    ],
    "footer": "{{ticket.created_at_iso | date: '%d %b %Y'}}"
}
```

Rows with an empty or unset `value` are omitted.

## Development

```sh
mise run test
mise run lint
mise run preview
```

`mise run preview` writes `preview.png` with a logo placeholder without contacting Freshservice or a printer. Set `LOGO_URL` to include deployment branding.
