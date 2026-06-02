FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files first
COPY go.mod ./

# Copy source code (needed for go mod tidy to resolve imports)
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
