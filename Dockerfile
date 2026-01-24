FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o hotlines3-api main.go

# Final stage
FROM alpine:latest

# Install dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Set timezone to Bangkok
ENV TZ=Asia/Bangkok

# Copy binary and config
COPY --from=builder /app/hotlines3-api .
COPY config.yaml .

EXPOSE 8080

# Run the application
CMD ["./hotlines3-api"]
