# Dockerfile for the gateway
FROM golang:1.23-alpine

WORKDIR /app

# Copy go.mod and go.sum to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build the application
RUN go build -o /app/gateway cmd/gateway/main.go

EXPOSE 8080

CMD ["/app/gateway"]