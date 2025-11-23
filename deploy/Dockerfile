# build stage
FROM golang:1.24.4 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# adjust path: build the package in ./cmd (you have cmd/main.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o pr-service ./cmd

# final stage
FROM alpine:3.20

# install ca-certificates in case you need TLS; remove if not needed
RUN apk add --no-cache ca-certificates

WORKDIR /app
# copy binary from builder
COPY --from=builder /app/pr-service .
# copy migrations from build context (ensure build context includes migrations/)
COPY migrations ./migrations

# create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown appuser:appgroup /app/pr-service /app/migrations

USER appuser

ENV APP_PORT=8080
EXPOSE 8080

CMD ["./pr-service"]