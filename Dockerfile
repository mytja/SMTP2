FROM golang:1.16-alpine AS builder

COPY . /app

WORKDIR /app

# Add gcc
RUN apk add build-base

RUN go mod download && \
    go env -w GOFLAGS=-mod=mod && \
    go build -v .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/SMTP2 ./SMTP2-server

EXPOSE 80
CMD [ "./SMTP2-server", "--useenv" ]
