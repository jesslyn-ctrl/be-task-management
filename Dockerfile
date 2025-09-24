# Stage 1: Builder
FROM golang:1.23.5-alpine as builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

RUN go build -o main ./cmd

# Stage 2: Runtime (Minimal)
FROM alpine:1.23.5

# Arguments
ARG SVC_PORT

WORKDIR /app

# Copy the built binary from builder stage
COPY --from=builder /app/main .

EXPOSE ${SVC_PORT}
CMD ["./main"]
