# =============================================================================
# Development Dockerfile — with Air hot reload
# =============================================================================
FROM golang:1.26.1-alpine3.23

RUN apk add --no-cache git curl make ca-certificates

RUN go install github.com/air-verse/air@latest

# Install Atlas CLI to a fixed path (atlasgo.sh installer is unreliable in Alpine).
RUN curl -fsSL "https://release.ariga.io/atlas/atlas-linux-amd64-latest" \
      -o /usr/local/bin/atlas \
    && chmod +x /usr/local/bin/atlas \
    && atlas version

RUN git config --global --add safe.directory /app

ENV GOFLAGS=-buildvcs=false

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
