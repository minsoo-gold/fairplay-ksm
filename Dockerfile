FROM golang:1.13-stretch AS builder

WORKDIR /app
COPY . /app

# Test
WORKDIR /app/ksm
RUN go get -u -t ./... && go test -run ''

WORKDIR /app/crypto
RUN go get -u -t ./... && go test -run ''

# Build
WORKDIR /app
RUN go build -o /app/ksm-server . && chmod 755 /app/ksm-server

FROM debian:stretch-slim
WORKDIR /app
COPY --from=builder /app/ksm-server .

EXPOSE 8080
ENTRYPOINT [ "/app/ksm-server" ]
