# freshservice-label

[![Release](https://img.shields.io/github/v/release/woodleighschool/freshservice-label?display_name=tag&sort=semver)](https://github.com/woodleighschool/freshservice-label/releases/latest)
[![CI](https://github.com/woodleighschool/freshservice-label/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/woodleighschool/freshservice-label/actions/workflows/ci.yaml)
[![Go](https://img.shields.io/github/go-mod/go-version/woodleighschool/freshservice-label?logo=go)](https://github.com/woodleighschool/freshservice-label/blob/main/go.mod)
[![Container](https://img.shields.io/badge/container-ghcr.io-2496ED?logo=github&logoColor=white)](https://github.com/orgs/woodleighschool/packages/container/package/freshservice-label)
[![License](https://img.shields.io/github/license/woodleighschool/freshservice-label)](https://github.com/woodleighschool/freshservice-label/blob/main/LICENSE)

Webhook service for printing Freshservice tickets on a Brother QL-820NWB. Labels use a fixed landscape layout for 62 mm continuous stock, and print jobs are queued in memory.

![Example label](example.png)

## 🚀 Usage

Create `.env` with the webhook secret and printer address:

```dotenv
FRESHSERVICE_LABEL_WEBHOOK_TOKEN=change-me
FRESHSERVICE_LABEL_PRINTER_ADDR=192.0.2.20
```

A container is published with each [release](https://github.com/woodleighschool/freshservice-label/releases/latest):

```bash
docker run --rm \
  --env-file .env \
  --publish 8080:8080 \
  ghcr.io/woodleighschool/freshservice-label:rolling
```

Send an Advanced JSON webhook from Freshservice to `/webhook` with `Authorization: Bearer <FRESHSERVICE_LABEL_WEBHOOK_TOKEN>`.

## ⚙️ Configuration

| Variable                           | Required | Default |
| ---------------------------------- | -------- | ------- |
| `FRESHSERVICE_LABEL_WEBHOOK_TOKEN` | Yes      |         |
| `FRESHSERVICE_LABEL_PRINTER_ADDR`  | Yes      |         |
| `FRESHSERVICE_LABEL_LOGO_URL`      | No       | No logo |
| `FRESHSERVICE_LABEL_LISTEN_ADDR`   | No       | `:8080` |
| `FRESHSERVICE_LABEL_QUEUE_DEPTH`   | No       | `10`    |
| `FRESHSERVICE_LABEL_PRINT_TIMEOUT` | No       | `30s`   |

When `FRESHSERVICE_LABEL_LOGO_URL` is set, the PNG is fetched once during startup and kept in memory.

## 🪝 Webhook

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

Rows with an empty or missing `value` are omitted.

## 🧑‍💻 Development

Run the current checkout with the same environment values:

```bash
FRESHSERVICE_LABEL_WEBHOOK_TOKEN=secret \
FRESHSERVICE_LABEL_PRINTER_ADDR=192.0.2.20 \
go run ./cmd/freshservice-label
```

Repository checks:

```bash
mise run test
mise run lint
mise run preview
```

`mise run preview` writes `preview.png` without contacting Freshservice or a printer. Set `FRESHSERVICE_LABEL_LOGO_URL` to render a logo.

## 📄 License

Licensed under the [Apache License 2.0](LICENSE).
