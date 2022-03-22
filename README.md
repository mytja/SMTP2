# SMTP2
![play_store_512](https://user-images.githubusercontent.com/52399966/149671395-c6a126a4-3c2b-48f3-9f00-64fef3214aad.png)

A Go implementation for SMTP2 protocol

# Received messages & sent messages
Sent Messages represent messages that were created on this server (host)
and later sent to a recipient (client) server.

Received messages represent all messages that were sent from
different server (host) and then arrived to this (client) server.

# Run it
How to run it (aka development phase):

Server 1: `go run . --host 127.0.0.1 --port 8080 --debug`

Server 2: `go run . --host 127.0.0.1 --port 8081 --dbconfig smtp2-1.db --debug`

# Special thanks
Thanks to following people for direct contribution:
- [Mitja](https://github.com/mytja)

Thanks to following people for direct support (contributing to issues and resolving many problems I encountered during this protocol journey):
- [Blaž Abram](https://github.com/BlazAbram)

Thanks to following people/organizations for indirect contribution (libraries):
- [Gorilla Web Toolkit](https://github.com/gorilla) for [mux](https://github.com/gorilla/mux)
- [golang-jwt contributors](https://github.com/golang-jwt) for [jwt](https://github.com/golang-jwt/jwt)
- [Jason Moiron](https://github.com/jmoiron) for [sqlx](https://github.com/jmoiron/sqlx)
- [Steve Francia](https://github.com/spf13) for [cobra](https://github.com/spf13/cobra)
- [Uber](https://go.uber.org) for [zap](https://go.uber.org/zap)
- [mattn](https://github.com/mattn) for [go-sqlite3](https://github.com/mattn/go-sqlite3)
- [Google Open Source](https://cs.opensource.google/) for [crypto](https://cs.opensource.google/go/x/crypto)

Thanks to following organizations and companies for their tools (indirect contribution):
- [GitHub](https://github.com) for providing [hosting](https://github.com) for this project and their amazing [GitHub Actions](https://github.com/features/actions), [GitHub Container Registry](https://ghcr.io) and [GitHub Issues](https://github.com/features/issues/). It really helped me while developing it (especially GitHub Issues and Actions)
- [JetBrains](https://jetbrains.com) for providing their awesome IDEs, like [GoLand](https://www.jetbrains.com/go/), which was used while developing this, and [PyCharm](https://www.jetbrains.com/pycharm/). Tools like Database inspector were so important for this project.

