# GoSight Agent Dockerfile (context: project root)

FROM golang:1.23.7 AS builder

WORKDIR /app
COPY agent/ ./agent/
COPY shared/ ./shared/
COPY go.work go.work.sum ./

WORKDIR /app/agent
RUN go mod download
RUN go build -o /gosight-agent ./cmd

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /gosight-agent /gosight-agent
COPY agent/config.yaml /config.yaml
COPY certs/ /certs/
ENTRYPOINT ["/gosight-agent"]
