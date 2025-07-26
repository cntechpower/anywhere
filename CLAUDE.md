# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Anywhere is a network tunneling tool similar to ngrok, written in Go with a React frontend. It provides TCP/UDP tunneling capabilities with enhanced visual management and high availability features. The system consists of server and agent components communicating via gRPC with TLS authentication.

## Key Components

- **Server**: Public-facing service that accepts incoming connections and forwards them to agents
- **Agent**: Runs on private networks and maintains persistent connections to the server
- **Frontend**: React-based web UI for configuration management and monitoring
- **gRPC**: Communication protocol between server and agents using Protocol Buffers

## Build Commands

### Go Backend
- `make build` - Build both server and agent binaries
- `make build_server` - Build server binary only
- `make build_agent` - Build agent binary only
- `make vet` - Run Go vet static analysis
- `make unittest` - Run Go unit tests
- `go test -count=1 -v ./...` - Run tests manually

### Frontend (React)
- `cd front-end && npm start` - Start development server
- `cd front-end && npm run build` - Build for production
- `cd front-end && npm test` - Run tests

### SSL Certificates
- `make newkey` - Generate new SSL certificates for TLS communication

### Protocol Buffers
- `make rpc` - Generate Go code from .proto files
- `make api` - Generate OpenAPI server code from YAML spec

## Architecture

### Server Architecture
- **main.go**: Entry point with Cobra CLI commands
- **server/**: Core server logic and proxy handling
- **api/**: REST API and gRPC endpoints for management
- **dao/**: Database access layer (SQLite)
- **model/**: Data models and structures

### Agent Architecture  
- **agent/**: Agent core logic and connection management
- **handler/**: gRPC handlers for agent commands
- **conn/**: Connection pooling and management

### Communication
- Server-Agent: gRPC over TLS (port 1111 default)
- Web UI: HTTPS (port 1114 default)
- REST API: HTTP (port 1112 default)
- Management gRPC: Internal (port 1113 default)

## Configuration

### Server Configuration
- `anywhered.json`: Main server configuration including ports, SSL settings, and user authentication
- Default admin credentials: admin/admin with OTP enabled
- TLS certificates in `credential/` directory

### Agent Configuration
- Command-line flags for server connection, zone assignment, and credentials
- Requires server IP, username, password, and zone name

## Development Workflow

1. Generate certificates: `make newkey`
2. Build binaries: `make build`
3. Start server: `./bin/anywhered start`
4. Start agent: `./bin/anywhere -i agent-1 -s SERVER_IP -u admin -z zone-1 --pass admin`
5. Access web UI: `https://localhost:1114`

## Testing

- Unit tests: `make unittest`
- Docker integration tests: `make docker_test`
- Frontend tests: `cd front-end && npm test`

## Files to Modify When

- **Protocol changes**: Update `.proto` files and run `make rpc`
- **API changes**: Update `anywhere.yml` and run `make api`
- **Frontend**: Work in `front-end/` directory
- **Server logic**: Modify files in `server/` directory
- **Agent logic**: Modify files in `agent/` directory