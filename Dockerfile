FROM golang:1.24.4

WORKDIR /app

COPY go.mod go.sum ./
COPY main.go ./
COPY handler/ ./handler

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /hello-http

EXPOSE 8080

