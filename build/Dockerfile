# =============================================================================
# Development Dockerfile — with Air hot reload
# =============================================================================
FROM golang:1.26.1-alpine3.23

RUN apk add --no-cache git curl make

RUN go install github.com/air-verse/air@latest
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN git config --global --add safe.directory /app

ENV GOFLAGS=-buildvcs=false

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
