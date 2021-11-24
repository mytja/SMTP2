FROM golang:1.16-alpine AS builder

COPY . /app

WORKDIR /app

RUN go mod download && \
    go build .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/SMTP2 ./SMTP2-server

EXPOSE 8080
CMD [ "./SMTP2-server --useenv" ]