# Build stage
FROM golang:alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reloader ./cmd/reloader

# Final stage
FROM alpine:3.19

# ca-certificates and curl are useful for making ExecHTTP calls if wget fails
RUN apk --no-cache add ca-certificates curl tzdata

WORKDIR /app
COPY --from=builder /app/reloader .

# Default port for metrics
EXPOSE 9090

ENTRYPOINT ["./reloader"]
