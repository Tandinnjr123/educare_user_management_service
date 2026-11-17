FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o usermanagementservice

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/usermanagementservice .
COPY --from=builder /app/.env .

EXPOSE 8093

HEALTHCHECK CMD wget -qO- http://localhost:8093/health || exit 1

CMD ["./usermanagementservice"]
