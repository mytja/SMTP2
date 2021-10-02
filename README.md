# SMTP2
A Go implementation for SMTP2 protocol

# Received messages & sent messages
Sent Messages represent messages that were created on this server (host)
and later sent to a recipient (client) server.

Received messages represent all messages that were sent from
different server (host) and then arrived to this (client) server.

# Run it
How to run it (aka development phase):

Server 1: `go run . --host 127.0.0.1 --port 8080`

Server 2: `go run . --host 127.0.0.1 --port 8081 --dbname smtp2-1.db`
