FROM golang:1.18-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/transaction-api ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/transaction-api .

EXPOSE 8080

CMD ["./transaction-api"]
