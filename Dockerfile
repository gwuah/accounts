FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o accounts cmd/accounts/main.go

FROM alpine:3.14
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/accounts .
ENV PORT=8080
EXPOSE $PORT
CMD ["./accounts"]