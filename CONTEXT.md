# Context: CCS-2PLUS PRO

## Project

CCS-2PLUS PRO is a hotel room management system for room status, check-in/check-out, maid or room-boy workflows, cashier operations, audit events, alerts, and manager dashboards.

The project is replacing a legacy CCS2 RS-485 hardware system with a newer Go backend, REST API, SSE realtime updates, and multiple Flutter apps.

## Current Architecture

Current active scope:

- `go-backend`: source of truth, business logic, room state, audit, alert, REST API, and SSE
- `flutter-dashboard`: dashboard, front office, manager view, reports
- `flutter-handheld`: handheld workflow for room boy or maid staff
- `flutter-cashier`: cashier and front-office workflow
- `frpc`: remote access and deployment tunnel

Topology:

```text
flutter-dashboard
flutter-handheld
flutter-cashier
        |
        | REST + SSE
        v
go-backend
```

PIC, ESP32 gateway, MQTT, and RS-485 transport are deferred unless the task explicitly says to work on them.

## Tech Stack

- Backend: Go 1.23, `net/http`
- Frontend: Flutter
- API: REST
- Realtime: Server-Sent Events at `GET /events`
- Database: no primary persistent database yet; room state is currently held in memory
- Deployment target: Raspberry Pi, `systemd`, `frpc`

## Important Commands

Backend:

```bash
go run ./go-backend
go test ./...
go build -o ccs2plus-backend ./go-backend
```

Backend with RS-485 hardware:

```bash
go run ./go-backend -listen :8080 -roommap path/to/roommap.json -serial /dev/ttyUSB0 -baud 9600
```

Flutter apps:

```bash
cd flutter-dashboard
flutter run --dart-define=BACKEND_BASE_URL=http://localhost:8080 --dart-define=PENDING_TIMEOUT_SECONDS=6
flutter analyze
flutter test
```

Use `flutter-handheld` or `flutter-cashier` instead of `flutter-dashboard` when working on those apps.

Mock backend scenarios:

```bash
./scripts/mock_backend_scenario.sh normal
./scripts/mock_backend_scenario.sh delayed-room-update
./scripts/mock_backend_scenario.sh sse-drop-after-full-state
```

## Important Docs

Read these before changing important behavior:

- `CLAUDE.md`: repository overview and primary commands
- `docs/current_architecture.md`: current architecture and active scope
- `docs/api_contract.md`: REST and SSE contract
- `docs/state_machine.md`: room status lifecycle
- `docs/db_model.md`: intended database model
- `docs/frontend_conventions.md`: Flutter conventions and checklist
- `docs/protocol.md`: legacy protocol reference
- `docs/rs485_transport_spec.md`: RS-485 reference
- `go-backend/roommap.example.json`: example room-to-controller mapping

## Domain Model

Important room state fields:

```json
{
  "room_id": "1205",
  "floor": 12,
  "status": "occupied",
  "stay_mode": "overnight",
  "case": "overnight",
  "remaining_minutes": 95,
  "controller_online": true
}
```

Main SSE events:

- `full_state`
- `room_update`
- `audit_event`
- `fraud_alert`
- `command_result`

Common room actions:

- `open_room`
- `service`
- `start_cleaning`
- `finish_cleaning`
- `check_in`
- `check_out`
- `extend_stay`

## Rules For Codex

- Before changing API behavior or room-state behavior, read `docs/api_contract.md` and `docs/state_machine.md`.
- Keep the REST and SSE contract stable unless the user explicitly asks to change it.
- Treat `go-backend` as the source of truth for room state and business logic.
- In Flutter apps, keep the existing pattern: config through `--dart-define`, SSE listener, local room-state map, and pending UI while waiting for `command_result`.
- Use memory mode when serial hardware is unavailable.
- Do not work on legacy RS-485, PIC, ESP32, or MQTT unless the task explicitly requires it.
- Keep frontend changes practical for real operational screens. This is not a marketing site.
- After backend changes, run `go test ./...` when feasible.
- After Flutter changes, run `flutter analyze` or `flutter test` when feasible.
- If tests or hardware verification cannot be run, say so clearly in the final response.

## Common Work Areas

- Backend API: `go-backend/internal/httpapi`
- Backend config: `go-backend/internal/config`
- Room map loading: `go-backend/internal/roommap`
- Transport and command writing: `go-backend/internal/transport`
- Legacy framing reference: `go-backend/internal/legacy`
- Dashboard UI: `flutter-dashboard`
- Handheld UI: `flutter-handheld`
- Cashier UI: `flutter-cashier`
- Deployment scripts: `scripts/`
- FRP config: `frpc/`

## Expected Output

When finishing a task, summarize:

- What files changed
- What behavior changed
- What tests or builds were run
- Any known limitations, skipped verification, or hardware assumptions
