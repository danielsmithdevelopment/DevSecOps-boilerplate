# Build stage
FROM golang:1.25.10-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o task-runner ./cmd/server

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /app/task-runner .

USER app
EXPOSE 8080
CMD ["./task-runner"]
