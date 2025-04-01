# GoSight Server Dockerfile
FROM golang:1.21 as builder

WORKDIR /app
COPY server/ ./server/
COPY shared/ ./shared/
COPY go.work go.work.sum ./

# Download dependencies
RUN cd server && go mod download

# Build server binary
RUN cd server && go build -o /gosight-server ./cmd

# Final image
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /gosight-server /gosight-server
COPY certs/ /certs/
COPY server/config.yaml /config.yaml
ENTRYPOINT ["/gosight-server"]
