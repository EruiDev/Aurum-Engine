FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o aurum ./cmd/api/api.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/aurum .

EXPOSE 8080

CMD ["./aurum"]