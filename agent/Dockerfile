# syntax=docker/dockerfile:1.4
FROM golang:1.23.7 AS builder

WORKDIR /app

# Copy Go workspaces
COPY agent/ ./agent/
COPY shared/ ./shared/
COPY go.work.docker go.work
COPY go.work.sum go.work.sum

# Build
WORKDIR /app/agent
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /gosight-agent ./cmd


# Runtime
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /gosight-agent /gosight-agent
COPY certs/ /certs/
COPY --chown=nonroot:nonroot certs/ /certs/
COPY --chown=nonroot:nonroot agent/config.docker.yaml /config.yaml

ENTRYPOINT ["/gosight-agent"]
