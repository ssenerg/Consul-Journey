# consul-journey

A small but production-shaped Go service that clusters itself entirely through
[HashiCorp Consul](https://www.consul.io/). Run several instances and they will:

- **Register** themselves with the local Consul agent (unique ID, tags, metadata).
- **Report health** two ways at once — an HTTP check Consul scrapes *and* a TTL
  check the app actively pushes live runtime metrics into.
- **Discover** every peer in real time via blocking queries on Consul's health API.
- **Elect a single leader** using a Consul session + KV lock, with leadership
  automatically released if a node becomes unhealthy or dies.

Kill the leader and a survivor takes over within seconds. Nothing but Consul is
used for coordination — no other database, no gossip code of our own.

## Which Consul features are used, and why

| Concern | Consul feature | Where |
| --- | --- | --- |
| Service registration | Agent service registration with tags + meta | `health.go: register` |
| Liveness (pull) | HTTP health check scraped by Consul | `health.go: register` |
| Liveness (push) | TTL check updated via `UpdateTTLOpts` with a JSON metrics snapshot | `health.go: runHeartbeat` |
| Self-healing catalog | `DeregisterCriticalServiceAfter` reaps dead instances | `health.go: register` |
| Peer discovery | `Health().Service` with **blocking queries** (watch, don't poll) | `discovery.go` |
| Leader election | **Session** (`Behavior: release`, `LockDelay`) + `KV.Acquire` lock | `election.go` |
| Health-gated leadership | Session bound to `serfHealth` + the TTL check | `election.go: createSession` |
| Leadership change watch | Blocking query on the leader KV key | `election.go: campaign` |
| Session liveness | `Session().RenewPeriodic` | `election.go: campaign` |
| Graceful shutdown | Destroy session + release lock + deregister on SIGТERM | `node.go: Run` |

### Why two health checks?

They cover different failure modes:

- The **HTTP check** (Consul → app) catches a crashed or unreachable process.
- The **TTL check** (app → Consul) catches a *wedged* process that still accepts
  TCP connections but has stopped making progress, and it carries app-computed
  health that only the app knows (goroutine count, heap, GC, leadership, peer
  count). It is the literal "send healthcheck data to Consul" path.

### Why session-based election is correct

Leadership is a Consul session holding a lock on one KV key. Because the session
is tied to the node's health checks, any of these releases leadership *without
the node having to notice*: the process dies, the node leaves the gossip pool,
or the TTL check lapses. `LockDelay` then briefly blocks re-acquisition to avoid
lock thrash during a partition. Exactly one node can hold the key at a time, so
there is exactly one leader.

## Running it

### 1. Start Consul

**Option A — Docker (default):**

```sh
docker compose -f deployment/docker/docker-compose.yml up -d   # Consul + UI on :8500
```

**Option B — native binary** (everything on `127.0.0.1`):

```sh
consul agent -dev -ui
```

> **Why the app needs an advertise address.** The app instances run on the host
> while Consul runs in a container, and Consul *actively scrapes* each instance's
> `/health`. From inside the container, `127.0.0.1` is the container itself — not
> the host — so the app must advertise a host-reachable address and bind all
> interfaces. The portable "the host" address is `host.docker.internal`
> (auto-injected by Docker Desktop; mapped via `extra_hosts` on Linux). Note that
> advertising `0.0.0.0` is rejected by Consul, so a concrete address is required.
>
> `run-cluster.sh` already defaults to `CJ_ADVERTISE_ADDR=host.docker.internal`
> and `CJ_BIND_ADDR=0.0.0.0`, so this is handled for you. When you run Consul as
> a **native binary** instead, override it: `CJ_ADVERTISE_ADDR=127.0.0.1 ./run-cluster.sh 3`.

### 2. Start a cluster of instances

```sh
./run-cluster.sh 3            # three instances on ports 8080, 8081, 8082
```

### 3. Watch it work

- Dashboards: <http://127.0.0.1:8080>, `:8081`, `:8082` (auto-refresh; the leader
  row is highlighted).
- Consul UI: <http://127.0.0.1:8500> — see the service, both health checks, the
  live TTL output, the session, and the leader KV key.
- Machine-readable status: `curl -s localhost:8080/status | jq`.

### 4. Test failover

Kill the current leader (find it on any dashboard or via `/status`) and watch a
new leader appear on the survivors within a couple of seconds:

```sh
curl -s localhost:8080/status | jq '{me:.node_id, leader:.leader.node_id}'
kill <leader-pid>             # or Ctrl-C the whole cluster
```

## Endpoints

| Path | Purpose |
| --- | --- |
| `GET /health` | Health check target for Consul; `200` passing, `503` critical. |
| `GET /status` | Full JSON: role, current leader, session, peers + their health. |
| `GET /` | Auto-refreshing HTML dashboard of the cluster as this node sees it. |

## Configuration

All flags have `CJ_*` environment fallbacks.

| Flag | Env | Default | Meaning |
| --- | --- | --- | --- |
| `-service` | `CJ_SERVICE` | `consul-journey` | Shared service name (discovery key). |
| `-id` | `CJ_NODE_ID` | `<host>-<port>` | Unique instance id. |
| `-consul` | `CJ_CONSUL_ADDR` | `127.0.0.1:8500` | Consul agent address. |
| `-bind` | `CJ_BIND_ADDR` | `127.0.0.1` | HTTP server bind address. |
| `-advertise` | `CJ_ADVERTISE_ADDR` | = bind | Address registered/scraped by Consul. |
| `-port` | `CJ_HTTP_PORT` | `8080` | HTTP API port. |
| `-dc` | `CJ_DATACENTER` | agent default | Consul datacenter. |
| `-leader-key` | `CJ_LEADER_KEY` | `service/<name>/leader` | KV key contended for leadership. |
