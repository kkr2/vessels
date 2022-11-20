FROM golang:1.17-buster as builder

# Create and change to the app directory.
WORKDIR /app
ENV config=docker
COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -v -o server ./cmd/main.go

FROM scratch
WORKDIR /
ENV config=docker
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/internal/config/* .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/server .
CMD ["./server"]

