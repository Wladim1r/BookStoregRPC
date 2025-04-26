### Build step
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/grpclib ./cmd/server/server.go

### Final step
FROM alpine:3.21

WORKDIR /grpc

COPY --from=builder /app/bin/grpclib .
COPY --from=builder /app/database /grps/database

CMD [ "./grpclib" ]