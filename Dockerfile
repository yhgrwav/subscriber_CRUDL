FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
COPY --from=builder /app/app /app/app
COPY --from=builder /app/docs /app/docs
RUN mkdir -p /app/logs && chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/app"]
