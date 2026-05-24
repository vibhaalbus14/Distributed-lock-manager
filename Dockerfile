# === STAGE 1: Low-Memory Build Stage ===
FROM golang:1.26-alpine AS builder 

# Disable CGO to keep the binary fully static and reduce build memory overhead
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# Copy dependency structures and download packages
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your backend files
COPY . .

# 🚀 FREE TIER TRICK: 
RUN go build -ldflags="-s -w" -p 2 -o dlm-server cmd/main.go

# === STAGE 2: Absolute Bare-Minimum Production Runtime ===
FROM scratch

WORKDIR /

# Copy only the compiled executable from the builder stage
COPY --from=builder /app/dlm-server .

# Expose your Gin port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./dlm-server"]