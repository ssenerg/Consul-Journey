set dotenv-load

module   := "consul-journey"
app-name := "Consul Journey"
version  := "v0.0.1"
revision := `\
    HASH=$(git rev-parse --short HEAD 2>/dev/null); \
    if [ -n "$(git status --porcelain 2>/dev/null)" ]; then \
        HASH="$HASH-dirty"; \
    fi; \
    echo $HASH \
`
ldflags  := "-s -w -X " + module + "/internal.module=" + module + " -X '" + module + "/internal.appName=" + app-name + "' -X " + module + "/internal.version=" + version + " -X " + module + "/internal.revision=" + revision

compose-location := "deployment/docker/docker-compose.yml"

alias d  := dev
alias p  := playground
alias c  := clean
alias pc := pre-commit
alias g  := generate

[private]
default:
    @echo "-----------------"
    @echo "{{ app-name }}"
    @echo "Version:  {{ version }}"
    @echo "Revision: {{ revision }}"
    @echo "-----------------"
    @just --list --unsorted

# Generate the protobuf code
[group('proto')]
generate:
    @mkdir -p proto/gen/go/node && \
        protoc -I proto \
        --go_out=proto/gen/go/node --go_opt=paths=source_relative \
        --go-grpc_out=proto/gen/go/node --go-grpc_opt=paths=source_relative \
        node.proto


# Run the application in dev mode
[group('go')]
dev instances="1": _build_server
    #!/usr/bin/env bash
    set -euo pipefail
    if ! [[ "{{instances}}" =~ ^[0-9]+$ ]] || [ "{{instances}}" -lt 1 ] || [ "{{instances}}" -gt 1000 ]; then
        echo "instances must be a positive integer between 1 and 1000 [inclusive] (got '{{instances}}')" >&2
        exit 1
    fi
    pids=()
    shutdown() {
        trap - INT TERM
        kill -TERM "${pids[@]}" 2>/dev/null || true
        wait 2>/dev/null || true
        exit 0
    }
    trap shutdown INT TERM
    for i in $(seq 1 {{instances}}); do
        name=$(printf "server-%03d" "$i")
        http_port=$((8080 + i - 1))
        grpc_port=$((9080 + i - 1))
        CJS_NODE_HTTP_CHECK_ADDRESS_OVERRIDE="host.docker.internal" \
        CJS_SERVER_HTTP_PORT="$http_port" \
        CJS_SERVER_GRPC_PORT="$grpc_port" \
        CJS_LOGGING_FILE_NAME="$name" \
        ./.bin/server > >(sed "s/^/[$name] /") 2>&1 &
        pids+=($!)
    done
    wait

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

# Clean the build binaries and logs
[group('go')]
clean:
    @rm -rf ./.bin
    @rm -rf ./.logs

# Run pre-commit checks
[group('go')]
pre-commit: _assert_build_flags tidy lint 

# Lint the codebase
[group('go')]
lint:
    @golangci-lint run --fix

# Tidy the dependencies
[group('go')]
tidy:
    @go mod tidy

# Build the server
_build_server: _assert_build_flags
    @go build -ldflags "{{ldflags}}" -o ./.bin/server ./cmd/server

# Ensure the build flags are valid
_assert_build_flags: _assert_version _assert_revision

# Ensure the version is a valid semver
_assert_version:
    @if ! echo "{{ version }}" | grep -E "^v(0|[1-9]\d*)\\.(0|[1-9]\d*)\\.(0|[1-9]\d*)$" >/dev/null; then \
        echo "version '{{ version }}' is not a valid semver" >&2; \
        exit 1; \
    fi

# Ensure the revision is a valid commit hash
_assert_revision:
    @if ! echo "{{ revision }}" | grep -E "^[a-f0-9]{7}(-dirty)?$" >/dev/null; then \
        echo "revision '{{ revision }}' is not a valid commit hash" >&2; \
        exit 1; \
    fi

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