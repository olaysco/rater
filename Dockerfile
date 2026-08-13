FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the API Gateway binary from cmd/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gateway ./cmd/gateway
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway .

EXPOSE 8080

CMD ["./gateway"]
