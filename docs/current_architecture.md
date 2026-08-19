# CCS-2PLUS PRO Current Architecture

เอกสารนี้สรุป architecture ปัจจุบันของโปรเจกต์แบบสั้นและชัด

---

## 1. Scope Now

ระบบในตอนนี้ประกอบด้วย:

- `go-backend`
- `flutter-dashboard`
- `flutter-handheld`
- `flutter-cashier`
- `frpc`

ยังไม่รวม:

- `PIC`
- `ESP32 Gateway`
- `MQTT`
- `RS-485 transport`

---

## 2. Topology

```text
flutter-dashboard
flutter-handheld
flutter-cashier
        |
        | REST + SSE
        v
go-backend
        |
        v
frpc
```

---

## 3. Roles

### go-backend

- source of truth
- business logic
- room state
- stay_mode / case
- audit / alert
- REST + SSE

### flutter-dashboard

- front office
- manager dashboard
- reports

### flutter-handheld

- mobile operator workflow
- room action commands

### flutter-cashier

- cashier / front office workflow
- room board and operator views

### frpc

- external access to backend/dashboard
- current public endpoint: `http://34.87.151.202:6010/`
- current remote SSH endpoint: `34.87.151.202:6110`
- production runtime on `raspi1` uses `systemd` for both `hotel-maid-backend` and `frpc`

---

## 4. Data Model Highlights

room state ที่ระบบควรถือ:

```json
{
  "room_id": "1205",
  "status": "occupied",
  "stay_mode": "overnight",
  "case": "overnight"
}
```

---

## 5. Deferred Legacy Layer

legacy hardware protocol ถูกพักไว้ก่อน แต่เอกสารยังเก็บอยู่ใน:

- [docs/protocol.md](/home/kaina/work/project/ccs2_plus/docs/protocol.md)
- [docs/rs485_transport_spec.md](/home/kaina/work/project/ccs2_plus/docs/rs485_transport_spec.md)
