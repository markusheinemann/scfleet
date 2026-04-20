# Stage 1: build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /agent ./cmd/agent

# Stage 2: runtime with Chromium
FROM alpine:3
RUN apk add --no-cache \
        ca-certificates \
        chromium \
        chromium-chromedriver \
        nss \
        freetype \
        harfbuzz \
        ttf-freefont \
    && ln -sf /usr/bin/chromium-browser /usr/bin/google-chrome
COPY --from=builder /agent /agent
ENTRYPOINT ["/agent"]
