# GoSight Agent Dockerfile
FROM golang:1.21 as builder

WORKDIR /app
COPY agent/ ./agent/
COPY shared/ ./shared/
COPY go.work go.work.sum ./

# Download dependencies
RUN cd agent && go mod download

# Build agent binary
RUN cd agent && go build -o /gosight-agent ./cmd

# Final image
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /gosight-agent /gosight-agent
COPY certs/ /certs/
COPY agent/config.yaml /config.yaml
ENTRYPOINT ["/gosight-agent"]
