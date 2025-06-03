FROM golang:1.24.3-alpine AS builder
WORKDIR /workspace
COPY . .
RUN go work sync
RUN go mod download
RUN go build -o /workspace/stats-service ./services/stats-service/cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /workspace/stats-service .
COPY services/stats-service/migrations ./migrations
EXPOSE 50052
CMD ["./stats-service"]
