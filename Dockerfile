# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies including Node.js for minification
RUN apk add --no-cache git nodejs npm

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install minification tools
RUN npm install -g terser clean-css-cli

# Minify JavaScript files
RUN terser public/js/clientLogin.js -o public/js/clientLogin.min.js --compress --mangle && \
    terser public/js/clientDashboard.js -o public/js/clientDashboard.min.js --compress --mangle && \
    terser public/js/statisticsModal.js -o public/js/statisticsModal.min.js --compress --mangle && \
    terser public/js/modalService.js -o public/js/modalService.min.js --compress --mangle && \
    terser public/js/bootstrap.js -o public/js/bootstrap.min.js --compress --mangle

# Minify CSS files
RUN cleancss -o public/css/bitaxeDashboard.min.css public/css/bitaxeDashboard.css && \
    cleancss -o public/css/modal.min.css public/css/modal.css && \
    cleancss -o public/css/statisticsModal.min.css public/css/statisticsModal.css && \
    cleancss -o public/css/bootstrap.min.css public/css/bootstrap.css

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o axeos-dashboard ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create app user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/axeos-dashboard .

# Copy public files from builder (includes minified files)
COPY --from=builder --chown=app:app /build/public ./public

# Copy config files into the image (can be overridden by volume mount)
COPY --chown=app:app config /app/config

# Switch to app user
USER app

# Expose port (tells Docker which port the container listens on)
EXPOSE 3000

# Set environment
ENV PORT=3000

# Add labels for VSCode Docker extension to auto-configure port mapping
# This makes right-click → Run work automatically on macOS
LABEL com.docker.extension.port.3000="3000"

# Health check - use public static file to avoid auth redirects in logs
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/public/css/bootstrap.min.css || exit 1

# Run the application
CMD ["./axeos-dashboard"]
