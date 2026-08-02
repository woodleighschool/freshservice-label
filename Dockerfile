# syntax=docker/dockerfile:1

# Defaults keep local builds self-contained. Renovate updates this alongside
# the matching Mise and module pins.
ARG GO_VERSION=1.26.5

# ---- Go build -------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache upx
WORKDIR /workspace

# Cache module downloads before copying source.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w" -o freshservice_label ./cmd/freshservice-label
RUN upx --best --lzma freshservice_label

# ---- Runtime --------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/freshservice_label /freshservice_label
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/freshservice_label"]
