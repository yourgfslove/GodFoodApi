FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/GodFoodApi/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .

COPY ./internal/sql/ ./sql


EXPOSE 8080
CMD ["./app"]


