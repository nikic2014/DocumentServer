FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /app/server .

RUN mkdir -p /app/uploads && chown -R appuser:appuser /app/uploads

USER appuser

EXPOSE 8080

ENTRYPOINT ["./server"]