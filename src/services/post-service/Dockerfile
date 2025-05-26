FROM golang:1.24.3-alpine AS builder
WORKDIR /workspace
COPY . .
RUN go work sync
RUN go mod download
RUN go build -o /workspace/post-service ./services/post-service/cmd

FROM alpine:latest
WORKDIR /app
COPY --from=builder /workspace/post-service .
COPY ./services/post-service/migrations ./migrations
EXPOSE 50051
CMD ["./post-service"]
