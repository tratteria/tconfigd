# First stage: build environment
FROM --platform=$BUILDPLATFORM golang:1.22.0 AS builder
ARG TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o tconfigd ./cmd

# Second stage: runtime environment
FROM --platform=$TARGETPLATFORM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/tconfigd .

LABEL org.opencontainers.image.source=https://github.com/tratteria/tconfigd
LABEL org.opencontainers.image.description="Tconfigd - Central daemon for distributing and managing Tratteria configurations"
LABEL org.opencontainers.image.licenses=Apache

ENTRYPOINT ["./tconfigd"]