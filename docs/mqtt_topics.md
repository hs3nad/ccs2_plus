# CCS-2PLUS PRO MQTT Topics

เอกสารนี้กำหนด MQTT contract สำหรับ `optional ESP32 gateway deployment`

หมายเหตุสำคัญ:

- deployment ปกติของระบบนี้ `ไม่ใช้ MQTT`
- โหมดปกติคือ `Raspberry Pi 1 -> RS-485 -> PIC slave/display`
- `MQTT` จะใช้เฉพาะเมื่อเปิดโหมด `ESP32 Gateway`
- `dashboard` และ `handheld/tablet` ปกติคุยกับ backend ผ่าน `REST + SSE`

---

## 1. When This Document Applies

เอกสารนี้ใช้เมื่อ topology เป็นแบบ:

```text
Raspberry Pi 1
  -> MQTT
  -> ESP32 Gateway
  -> RS-485
  -> PIC slave / PIC display
```

ถ้าเป็นโหมดปกติ:

```text
Raspberry Pi 1
  -> RS-485
  -> PIC slave / PIC display
```

ให้ข้ามเอกสาร MQTT นี้ได้

---

## 2. Topic Design Principles

- backend เป็นผู้ตีความ state หลัก
- topic ตั้งชื่อแบบ `ccs2plus/{domain}/{target}/{action}`
- command และ ack ต้องมี `request_id`
- snapshot/state topic ที่ใช้โหลดค่าเริ่มต้นควร retained
- event log และ alert ไม่ควร retained

---

## 3. Topic Overview

| Topic | Direction | QoS | Retained | Purpose |
|---|---|---|---|---|
| `ccs2plus/backend/state/full` | Backend -> Clients | 1 | yes | full snapshot |
| `ccs2plus/backend/state/room/{room_id}` | Backend -> Clients | 1 | yes | room state ล่าสุด |
| `ccs2plus/backend/event/room` | Backend -> Clients | 0 | no | realtime room update |
| `ccs2plus/backend/event/audit` | Backend -> Clients | 0 | no | audit trail event |
| `ccs2plus/backend/event/alert` | Backend -> Clients | 0 | no | anti-fraud / system alert |
| `ccs2plus/backend/system/status` | Backend -> Clients | 1 | yes | backend status |
| `ccs2plus/client/request/full_state` | Client -> Backend | 0 | no | ask backend to republish snapshot |
| `ccs2plus/client/command/room` | Client -> Backend | 0 | no | room action from dashboard/handheld |
| `ccs2plus/client/command/checkin` | Client -> Backend | 0 | no | check-in / check-out / extend |
| `ccs2plus/gateway/{gateway_id}/hello` | Gateway -> Backend | 1 | no | announce gateway |
| `ccs2plus/gateway/{gateway_id}/heartbeat` | Gateway -> Backend | 0 | no | health heartbeat |
| `ccs2plus/gateway/{gateway_id}/snapshot` | Gateway -> Backend | 0 | no | full floor/device state |
| `ccs2plus/gateway/{gateway_id}/event` | Gateway -> Backend | 0 | no | device event from PIC |
| `ccs2plus/gateway/{gateway_id}/command` | Backend -> Gateway | 1 | no | command to gateway |
| `ccs2plus/gateway/{gateway_id}/command_ack` | Gateway -> Backend | 1 | no | ack/result of command |

---

## 4. Client To Backend Payloads

### 3.1 Room Command

Topic: `ccs2plus/client/command/room`

```json
{
  "request_id": "01JV7QTHJZ3X8P8Y9V7M7A0M2A",
  "timestamp": "2026-05-11T20:15:00+07:00",
  "source": "handheld",
  "actor": {
    "user_id": "roomboy02",
    "device_id": "HH-03"
  },
  "room_id": "1205",
  "action": "open_room",
  "reason": "guest_request",
  "meta": {
    "gps": {
      "lat": 13.7563,
      "lng": 100.5018
    }
  }
}
```

Supported `action`:

- `open_room`
- `close_room`
- `start_cleaning`
- `finish_cleaning`
- `set_dnd`
- `clear_dnd`
- `set_makeup`
- `set_urgent_makeup`

### 3.2 Check-In Command

Topic: `ccs2plus/client/command/checkin`

```json
{
  "request_id": "01JV7QVC6R1AFHK4WJ9D8J4BEP",
  "timestamp": "2026-05-11T20:16:00+07:00",
  "source": "dashboard",
  "actor": {
    "user_id": "cashier01"
  },
  "room_id": "1205",
  "action": "check_in",
  "stay_hours": 2,
  "payment_amount": 300.0,
  "payment_ref": "POS-20260511-0001"
}
```

Supported `action`:

- `check_in`
- `check_out`
- `extend_stay`

---

## 5. Backend To Client Payloads

### 4.1 Full State

Topic: `ccs2plus/backend/state/full`

```json
{
  "type": "full_state",
  "timestamp": "2026-05-11T20:17:00+07:00",
  "rooms": [
    {
      "room_id": "1205",
      "floor": 12,
      "status": "occupied",
      "checkin_state": "in_use",
      "controller_online": true,
      "guest_name": null,
      "remaining_minutes": 95,
      "flags": {
        "dnd": false,
        "makeup": false,
        "urgent_makeup": false,
        "cleaning": false
      }
    }
  ],
  "summary": {
    "vacant": 32,
    "occupied": 12,
    "cleaning": 4,
    "dnd": 2,
    "overstay": 1
  }
}
```

### 4.2 Room Update

Topic: `ccs2plus/backend/event/room`

```json
{
  "type": "room_update",
  "timestamp": "2026-05-11T20:17:15+07:00",
  "request_id": "01JV7QTHJZ3X8P8Y9V7M7A0M2A",
  "room_id": "1205",
  "status": "cleaning",
  "actual_state": {
    "door_open": true,
    "guest_power": false,
    "maid_mode": true
  },
  "actor": {
    "user_id": "roomboy02",
    "device_id": "HH-03"
  }
}
```

### 4.3 Audit Event

Topic: `ccs2plus/backend/event/audit`

```json
{
  "type": "audit_event",
  "timestamp": "2026-05-11T20:17:15+07:00",
  "request_id": "01JV7QTHJZ3X8P8Y9V7M7A0M2A",
  "room_id": "1205",
  "action": "open_room",
  "result": "success",
  "actor": {
    "user_id": "roomboy02",
    "device_id": "HH-03"
  },
  "source": "handheld",
  "trace": {
    "gateway_id": "floor-12-a",
    "controller_id": "ctrl-1205"
  }
}
```

### 4.4 Alert Event

Topic: `ccs2plus/backend/event/alert`

```json
{
  "type": "fraud_alert",
  "timestamp": "2026-05-11T20:20:00+07:00",
  "room_id": "1205",
  "severity": "high",
  "rule_code": "UNAUTHORIZED_ROOM_POWER_ON",
  "message": "Room power turned on without check-in",
  "actual_state": {
    "guest_power": true
  }
}
```

---

## 6. Gateway Payloads

### 5.1 Gateway Hello

Topic: `ccs2plus/gateway/{gateway_id}/hello`

```json
{
  "gateway_id": "floor-12-a",
  "timestamp": "2026-05-11T20:10:00+07:00",
  "floor": 12,
  "firmware_version": "0.1.0",
  "protocol": "rs485-modbus-rtu",
  "room_range": ["1201", "1216"]
}
```

### 5.2 Gateway Event

Topic: `ccs2plus/gateway/{gateway_id}/event`

```json
{
  "timestamp": "2026-05-11T20:17:14+07:00",
  "room_id": "1205",
  "controller_id": "ctrl-1205",
  "event_type": "sensor_update",
  "register": "0x14",
  "raw": {
    "inpbuf0": 8,
    "outbuf4": 64
  },
  "actual_state": {
    "guest": false,
    "maid": true,
    "dnd": false,
    "door_open": true
  }
}
```

### 5.3 Gateway Command

Topic: `ccs2plus/gateway/{gateway_id}/command`

```json
{
  "request_id": "01JV7QTHJZ3X8P8Y9V7M7A0M2A",
  "timestamp": "2026-05-11T20:17:13+07:00",
  "room_id": "1205",
  "controller_id": "ctrl-1205",
  "command": "write_register",
  "register": "0x14",
  "value": 64,
  "mask": 64
}
```

### 5.4 Gateway Command ACK

Topic: `ccs2plus/gateway/{gateway_id}/command_ack`

```json
{
  "request_id": "01JV7QTHJZ3X8P8Y9V7M7A0M2A",
  "timestamp": "2026-05-11T20:17:14+07:00",
  "room_id": "1205",
  "result": "success",
  "ack": true,
  "actual_state": {
    "maid": true,
    "guest": false
  },
  "latency_ms": 420
}
```

---

## 7. Delivery Rules

- ถ้า backend ได้ command แต่ไม่เห็น `command_ack` ภายใน timeout ให้ mark `failed`
- backend จะ publish `room_update` หลัง ACK สำเร็จเท่านั้น
- handheld และ dashboard ไม่ควรตีความ hardware state โดยตรงจาก gateway topic
- client ฝั่ง UI ควรใช้ `backend/state/*` และ `backend/event/*` เป็นหลัก
