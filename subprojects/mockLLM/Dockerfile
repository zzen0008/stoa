# Dockerfile for the mockLLM server
FROM golang:1.23-alpine

WORKDIR /app

# Copy only the go.mod file first to leverage Docker cache
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY main.go ./

# Build the application
RUN go build -o /app/mockLLM main.go

EXPOSE 8081

CMD ["/app/mockLLM"]