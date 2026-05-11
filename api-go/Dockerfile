FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /api-go ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /api-go .
EXPOSE 8080
CMD ["./api-go"]
