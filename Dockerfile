FROM 17-alpine3.12 AS builder

COPY . /app

WORKDIR /app

RUN go mod download && \
    go env -w GOFLAGS=-mod=mod && \
    go build .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/SMTP2 ./SMTP2-server

EXPOSE 8080
CMD [ "./SMTP2-server --useenv" ]