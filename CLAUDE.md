# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Anywhere is a network tunneling tool similar to ngrok, written in Go with a React frontend. It provides TCP/UDP
tunneling with visual management and high availability. Server and agent communicate over mutual TLS with a
length-prefixed JSON message protocol.

## Key Components

- **Server**: public-facing service that accepts agent control connections, opens proxy listeners and forwards
  client traffic to agents (`server/`).
- **Agent**: runs in the private network, keeps one control connection to the server and dials local services on
  demand (`agent/`).
- **Frontend**: React admin UI served by the server from `./static` (`front-end/`).
- **Protocol**: JSON messages framed as `[4-byte little-endian length][payload]`, see `conn/conn.go`.

## Build Commands

```shell
make build          # bin/anywhered + bin/anywhere (uses local go, no docker)
make build_server   # server only
make build_agent    # agent only
make test           # go test ./...
make race           # go test -race ./...
make vet            # go vet ./...
make fmt            # gofmt hand written Go files
make newkey         # regenerate CA/server/client/http certs (SAN: anywhere.local, localhost, 127.0.0.1)
make rpc            # regenerate protobuf code (needs protoc + plugins)
make api            # regenerate swagger server code (needs go-swagger v0.30.5)
make ui             # npm ci + vite build into ./static
make build_linux    # Linux binaries for the container E2E (golang:1.22 in Docker)
make docker_test    # container E2E: 3 agents + HTTP/UDP/MySQL/console (needs Docker)
make docker_test_clean
cd front-end && npm run dev   # Vite dev server on :3000 (VITE_API_PREFIX optional)
cd front-end && npm run lint  # ESLint (TypeScript)
cd front-end && npm test      # Vitest + jsdom
make clean
```

Toolchain: Go 1.22+, Node 20+, OpenSSL 1.1.1+.

## Ports

| Service | Protocol | Default |
|---------|----------|---------|
| Server-Agent | gRPC-like TCP over mTLS | 1111 |
| REST API | HTTP(S) | 1114 (`/api` prefix) |
| Management gRPC | mTLS, loopback | 1113 |
| Agent admin gRPC | plaintext, loopback only | 1110 |

## Architecture

### Connection model

1. Agent dials the server over mTLS and sends `PkgControlConnRegister` (username/password).
2. The server registers the agent in a zone and replies to heartbeats every second.
3. When a client connects to a proxy port, the server asks a healthy agent for a tunnel
   (`AskProxyConn`). The request carries a **single use token**.
4. The agent dials the local service and opens a second TLS connection to the server with
   `PkgTunnelBegin` + the token. The server validates the token, then puts the connection into the
   per-`localAddr` pool (`conn/pool.go`).
5. Client and agent connections are bridged with `conn.JoinConnWithIdleTimeout`.

### Server data structures

- There are no package level server/agent/handler singletons: `server.New`, `agent.New`,
  `handler.New` and `handler.NewRESTHandler` take their dependencies explicitly.
- `server/server.Server.zones` is guarded by `zonesMu`; always use `getZone`/`getOrCreateZone`,
  never index the maps directly.
- `server/zone.Zone` guards `agents` and `proxyConfigs` with their own RWMutex. Listeners are bound
  synchronously in `AddProxyConfig` so a port conflict is reported to the caller.
- `ProxyConfigStats` keeps counters in atomics and flushes them to SQLite periodically
  (`FlushStats`); do not write to the database per connection.
- `dao/connlist.JoinedConnList` is in-memory only and uses globally unique IDs.

### Protocol / protocol changes

- Framing lives in `conn/conn.go` (`writeFrame`/`readFrame`), with a 16MiB max frame and full
  `io.ReadFull` semantics. Never reintroduce a per-call `bufio.Reader` or rely on `Buffered()`.
- Message types are in `model/package.go`; a data connection without a valid token must be rejected.
- UDP uses `PkgUDPData` (server -> agent) and `PkgUDPDataResp` (agent -> server) with a `SessionID`
  that maps back to the client endpoint.

### Auth and TLS

- UI login requires username + password and, when enabled, TOTP (`server/api/auth`). Passwords may be
  bcrypt (`user_password_hash`) or legacy plaintext.
- Session tokens are HS256 JWTs signed with `ui_config.jwt_secret`; sessions are HttpOnly,
  SameSite=Lax cookies, and state changing requests are Origin checked.
- Server TLS requires client certificates. Clients verify the server certificate; the server name
  must match a SAN (`--tls-server-name`, default `anywhere.local`). `--tls-insecure` is a migration
  escape hatch only.
- `internal/trace` wraps opentracing so tracing is a no-op when `TRACE_ADDR` is unset (the upstream
  `utils/tracing.New` panics without `Init`).

## Configuration

- `anywhered.json` is created by `./bin/anywhered config reset`, which prints random admin
  credentials (password + OTP secret) once.
- `ui_config`: `skip_login`, `jwt_secret`, `grpc_addr`, `grpc_insecure`, `web_ui_listen_addr`,
  `debug`, `enable_tls`.
- `proxy_conn_idle_timeout_seconds`: closes fully idle proxied TCP connections (0 = disabled,
  which is what long lived SSH tunnels need).
- `ssl_config` (agent facing) and `http_ssl_config` (web UI) point at certificate files.
- `user_config.users[]`: `user_name`, `user_password` **or** `user_password_hash`, `otp_enable`,
  `otp_code`.
- Agent flags: `--server-ip/-s`, `--server-port/-p`, `--zone-name/-z`, `--agent-id/-i`,
  `--user/-u`, `--pass`, `--cert`, `--key`, `--ca`, `--tls-server-name`, `--tls-insecure`,
  `--grpc-address`.

## Frontend

- Toolchain: Vite 8 + React 17 + TypeScript 5 + Ant Design 4; Vitest 5. Entry: `src/main.tsx`,
  config: `vite.config.ts` (output `build/`), tests: Vitest + jsdom.
- Design system: `design-system/anywhere/MASTER.md` (dark operations console, Fira Sans/Code,
  green accent, glass panels). Keep new pages consistent with it.
- Tokens and shell styles: `front-end/src/theme.css`; Ant Design 4 dark overrides:
  `front-end/src/antd-dark.css`. Import order in `src/main.tsx` matters:
  `antd.css` -> `index.css` -> `theme.css` -> `antd-dark.css`.
- Shared components live in `front-end/src/components` (App shell, PageHeader, StatCard,
  StatusBadge, BarList, states, WhiteListFields); API types in `src/types.ts`.
- Data loading: `src/hooks/useApi.ts` (polling + abort + stale data). HTTP: `src/api/client.ts`
  (session cookies, 401 redirect). API base URL: `VITE_API_PREFIX` (empty = same origin).
- Responsive: sidebar becomes a drawer below 960px, tables switch to cards below 860px.

## Testing

- `make test` / `make race` run all unit tests.
- `server/server/e2e_test.go` starts a real TLS server + agent and proxies a TCP echo, covering the
  whole control/data path including token validation.
- `server/zone/zone_test.go` covers listener binding, token binding and UDP echo.
- Tests use an in-memory SQLite database via `ANYWHERE_DB_DSN`; never point tests at `proxy.db`.
- `make docker_test` runs `test/e2e/run.sh`: builds Linux binaries/images, starts compose targets,
  logs in through the HTTP API, creates TCP/UDP/MySQL proxies and verifies each data path plus the
  console assets. `KEEP=1` keeps the project running for debugging; `SKIP_UI=1` reuses `./static`.

## Dependency notes

- Swagger server code is generated with go-swagger v0.30.5 (`make api`). The go-openapi modules are
  pinned to the generator's versions (errors 0.20, runtime 0.26, spec 0.20, ...). Regenerate the
  code when bumping them, and keep `--name AnywhereServer` so the API type is not renamed.
  Other dependencies (gin, gorm, jwt, cobra, tablewriter, otp) are on current versions.

## Files to Modify When

- **Protocol changes**: update `model/package.go` and, for RPC, the `.proto` files + `make rpc`.
- **API changes**: update `server/api/http/definition/anywhere.yml` and run `make api`.
- **Frontend**: `front-end/` (`npm run dev`, `npm run build`, `npm run lint`). API base URL comes from
  `VITE_API_PREFIX` (empty = same origin).
- **Server logic**: `server/server/` (lifecycle, summary) and `server/zone/` (proxy handling).
- **Agent logic**: `agent/agent/` (control loop, tunnels, UDP sessions).
