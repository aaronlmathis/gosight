FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o gosight ./cmd
EXPOSE 8080
CMD ["./gosight"]
