FROM golang:1.24.3-alpine AS builder
WORKDIR /workspace
COPY go.work go.work.sum ./
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY services/post-service/go.mod services/post-service/go.sum ./services/post-service/
COPY services/stats-service/go.mod services/stats-service/go.sum ./services/stats-service/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
RUN go work sync
COPY . .
RUN go build -o /app/stats-service ./services/stats-service/cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/stats-service .
COPY ./services/stats-service/migrations ./migrations
EXPOSE 50052
CMD ["./stats-service"]