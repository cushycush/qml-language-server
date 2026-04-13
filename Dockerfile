FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o qml-language-server .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/qml-language-server .

ENV PATH="/app:${PATH}"

ENTRYPOINT ["/app/qml-language-server"]

LABEL org.opencontainers.image.title="QML Language Server"
LABEL org.opencontainers.image.description="A Go-based Language Server for QML"
LABEL org.opencontainers.image.source="https://github.com/cushycush/qml-language-server"
LABEL org.opencontainers.image.licenses="MIT"
