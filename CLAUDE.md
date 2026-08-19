# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

**CCS-2PLUS PRO** is a hotel room management system. It replaces a legacy RS-485 hardware system (CCS2) with a modern REST + SSE architecture. The backend is a single Go server; the frontend is three separate Flutter apps targeting different staff roles.

## Commands

### Go Backend

```bash
# Run in memory mode (no hardware required)
go run ./go-backend

# Run with RS-485 hardware
go run ./go-backend -listen :8080 -roommap path/to/roommap.json -serial /dev/ttyUSB0 -baud 9600

# Build binary
go build -o ccs2plus-backend ./go-backend

# Run all tests
go test ./...

# Run a single test
go test ./go-backend/internal/httpapi/ -run TestCheckIn
```

### Flutter Apps (dashboard / handheld / cashier)

```bash
# Run (substitute the correct app directory)
cd flutter-dashboard   # or flutter-handheld / flutter-cashier
flutter run \
  --dart-define=BACKEND_BASE_URL=http://localhost:8080 \
  --dart-define=PENDING_TIMEOUT_SECONDS=6

# Lint
flutter analyze

# Run widget tests
flutter test
```

### Mock Backend Scenarios

```bash
# Simulate different backend behaviors for frontend testing
./scripts/mock_backend_scenario.sh normal
./scripts/mock_backend_scenario.sh delayed-room-update
./scripts/mock_backend_scenario.sh sse-drop-after-full-state
```

## Architecture

```
flutter-dashboard  ─┐
flutter-handheld   ─┤─ REST + SSE ─ go-backend ─ RS-485 (via serial)
flutter-cashier    ─┘
```

### go-backend

Single HTTP server (net/http). Package layout under `go-backend/internal/`:

| Package | Responsibility |
|---|---|
| `config` | CLI flag parsing (listen addr, roommap path, serial, baud, write delay) |
| `httpapi` | Route registration and HTTP handlers |
| `transport` | Translates REST commands into RS-485 frames; `MemoryWriter` used when no serial device is configured |
| `roommap` | Loads `roommap.json` — maps room IDs to controller addresses/ports |
| `legacy` | RS-485 protocol framing (deferred; not yet wired into main data path) |

**SSE stream** (`GET /events`) is the primary real-time channel. Clients connect on startup and receive:
- `full_state` — complete snapshot of all rooms
- `room_update` — single-room delta
- `audit_event` / `fraud_alert` / `command_result`

**Room state** is held in memory (no persistent DB yet). Key fields: `status`, `stay_mode`, `case`, `remaining_minutes`, `controller_online`.

### Flutter Apps

All three apps share the same pattern:
- `lib/src/config/app_config.dart` — reads `--dart-define` values
- An SSE listener that maintains a local room state map, updated by `room_update` events
- REST calls for actions; a "pending" UI state is shown while awaiting the `command_result` event (timeout controlled by `PENDING_TIMEOUT_SECONDS`)

**Role split:**
- **dashboard** — manager/front-office overview, audit log, fraud alerts, reports
- **handheld** — room-boy/maid workflow: open_room, start/finish cleaning, service
- **cashier** — check-in, check-out, extend stay, receive payment

## Key Reference Docs

- [docs/api_contract.md](docs/api_contract.md) — full REST + SSE API specification
- [docs/state_machine.md](docs/state_machine.md) — room status lifecycle (vacant → occupied → cleaning → …)
- [docs/db_model.md](docs/db_model.md) — intended schema (rooms, stays, audit_events)
- [docs/frontend_conventions.md](docs/frontend_conventions.md) — Flutter coding rules and pre-merge checklist
- [docs/current_architecture.md](docs/current_architecture.md) — narrative architecture overview
- [roommap.example.json](go-backend/roommap.example.json) — example room-to-controller mapping