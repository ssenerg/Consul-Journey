set dotenv-load

compose-location := "deployment/docker/docker-compose.yml"

alias r := run
alias p := playground

[private]
default:
    @just --list --unsorted

# Run the application
[group('go')]
run: _build
    @./.bin/main

# Run the playground
[group('go')]
playground:
    @if [ ! -d playground ]; then \
        mkdir playground; \
    fi
    @if [ -z "$(ls playground/*.go 2>/dev/null)" ]; then \
        echo 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello from Playground")\n}' > playground/main.go; \
    fi
    @go build -o ./.bin/playground ./playground
    @./.bin/playground

_build:
    @go build -o ./.bin/main main.go

# Compose up
[group('compose')]
up service="":
    @docker compose -f {{compose-location}} up -d {{service}}

# Compose down
[group('compose')]
down service="":
    @docker compose -f {{compose-location}} down {{service}}

# Compose down with volumes
[group('compose')]
downv service="":
    @docker compose -f {{compose-location}} down -v {{service}}

# Compose logs
[group('compose')]
logs service="":
    @docker compose -f {{compose-location}} logs -f {{service}}

# Compose restart
[group('compose')]
restart service="":
    @docker compose -f {{compose-location}} restart {{service}}