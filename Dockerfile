# builder
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -o /server ./api

FROM gcr.io/distroless/base-debian12
COPY --from=builder /server /server
USER nonroot:nonroot
ENV PORT=8080
EXPOSE 8080
CMD ["/server"]

# 디버그 (Alpine)
#FROM alpine:3.20 AS debug
#RUN apk add --no-cache bash curl
#COPY --from=builder /server /server
#ENV PORT=8080
#EXPOSE 8080
#CMD ["/server"]