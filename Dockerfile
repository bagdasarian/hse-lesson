FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG CMD=cmd/main.go
RUN go build -o server ./${CMD}

FROM alpine:3.23.4
WORKDIR /app
COPY --from=builder /app/server .
CMD ["./server"]
