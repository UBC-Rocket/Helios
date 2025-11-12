# Builder
FROM golang:1.25.2 AS builder

WORKDIR /app

COPY /internal /app/internal

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/helios ./cmd/helios

# Runtime
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/ /app/bin

EXPOSE 8080

CMD ["./bin/helios"]