FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/task-management ./cmd/app

FROM alpine:3.20

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /root

COPY --from=builder /app/task-management .

EXPOSE 8080

ENTRYPOINT ["./task-management"]
