FROM golang:1.24.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o post-service ./cmd

FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/post-service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 50051
CMD ["./post-service"]
