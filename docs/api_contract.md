# CCS-2PLUS PRO API Contract

เอกสารนี้กำหนด REST API และ SSE สำหรับ `dashboard` และ `handheld/tablet`

---

## 1. API Design Rules

- REST ใช้สำหรับ query และ command จาก UI
- SSE ใช้สำหรับ realtime updates
- backend ตอบรับคำสั่งเป็น `accepted` หรือ `success/failure`
- action ที่เกี่ยวกับ hardware ต้องอิง `request_id`
- ทุก response ใช้เวลาแบบ ISO-8601

---

## 2. Health And Session

### `GET /ping`

Response:

```json
{
  "status": "ok",
  "time": "2026-05-11T20:30:00+07:00"
}
```

### `GET /api/me`

Response:

```json
{
  "user_id": "roomboy02",
  "role": "handheld",
  "device_id": "HH-03",
  "permissions": ["open_room", "start_cleaning", "finish_cleaning"]
}
```

---

## 3. State Read APIs

### `GET /api/state`

ใช้สำหรับ dashboard โหลด snapshot ตอนเริ่มต้น

```json
{
  "rooms": [
    {
      "room_id": "1205",
      "floor": 12,
      "status": "occupied",
      "stay_mode": "overnight",
      "case": "overnight",
      "remaining_minutes": 95,
      "controller_online": true
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

### `GET /api/rooms/{room_id}`

```json
{
  "room_id": "1205",
  "status": "occupied",
  "stay_mode": "overnight",
  "case": "overnight",
  "checkin_state": "in_use",
  "controller_online": true,
  "flags": {
    "dnd": false,
    "makeup": false,
    "cleaning": false
  },
  "actual_state": {
    "door_open": false,
    "guest_power": true
  }
}
```

### `GET /api/handheld/tasks`

รายการห้องหรือ assignment ของพนักงาน

```json
{
  "user_id": "roomboy02",
  "rooms": ["1201", "1205", "1208"]
}
```

---

## 4. Room Command APIs

### `POST /api/rooms/{room_id}/command`

Request:

```json
{
  "request_id": "01JV7R1H2CNH04H8J01KP6X1QV",
  "action": "open_room",
  "source": "handheld",
  "device_id": "HH-03",
  "reason": "guest_request",
  "gps": {
    "lat": 13.7563,
    "lng": 100.5018
  }
}
```

Response:

```json
{
  "request_id": "01JV7R1H2CNH04H8J01KP6X1QV",
  "status": "accepted",
  "room_id": "1205"
}
```

Supported actions:

- `open_room`
- `service`
- `start_cleaning`
- `finish_cleaning`

หมายเหตุ:

- `open_room` และ `service` ใน backend ปัจจุบันเป็น software workflow command
- ทั้งสอง action จะถูกบันทึกลง `audit` และใช้สร้าง `alerts` บางกรณี
- ยังไม่ map ลง legacy transport โดยตรงใน scope ปัจจุบัน

### `POST /api/rooms/{room_id}/checkin`

Request:

```json
{
  "request_id": "01JV7R3B0PSJ3MZQAR0KSB4QXR",
  "action": "check_in",
  "stay_mode": "temporary",
  "case": "temporary",
  "stay_hours": 2,
  "payment_amount": 300.0,
  "payment_ref": "POS-20260511-0001"
}
```

### `POST /api/rooms/{room_id}/checkout`

Request:

```json
{
  "request_id": "01JV7R3YQF8N2K0PZQ9QCTW3Y5",
  "action": "check_out"
}
```

### `POST /api/rooms/{room_id}/extend`

Request:

```json
{
  "request_id": "01JV7R4CQA0J4E55F27F6VM1VQ",
  "action": "extend_stay",
  "stay_mode": "temporary",
  "case": "temporary",
  "extend_minutes": 60,
  "payment_amount": 150.0,
  "payment_ref": "POS-20260511-0002"
}
```

### `POST /api/rooms/{room_id}/payment`

Request:

```json
{
  "request_id": "01JV7R4ZZPAY0000000000001",
  "action": "receive_payment",
  "amount": 150.0,
  "method": "Cash",
  "reference": "PAY-20260511-0003"
}
```

Supported `stay_mode` / `case` values:

- `overnight`
- `temporary`
- `monthly`
- `yearly`

หมายเหตุ:

- `stay_mode` ใช้สื่อ semantic mode ของการเข้าพัก
- `case` เก็บค่าเพื่อ compatibility กับ report เดิมของ CCS2
- ในหลายระบบใหม่สามารถให้ `case` มีค่าเท่ากับ `stay_mode` ไปก่อน
- สำหรับ `AMP_V2D` generation ฝั่ง display:
  - `start_cleaning` ใช้ `LED_MODE = 3` (`on`) ตอนเปลี่ยนสถานะหลัก
  - `overstay` ใช้ `LED_MODE = 3` (`on`) ตอนเข้าภาวะ overtime หลัก
  - `LED_MODE = 1/2` (`slow/fast blink`) เป็น warning phase จาก timer logic ไม่ใช่ default mapping ของ state หลัก

---

## 5. Audit And Alert APIs

### `GET /api/audit/events`

Query examples:

- `?room_id=1205`
- `?request_id=01JV7R1H2CNH04H8J01KP6X1QV`
- `?limit=50`

Response:

```json
{
  "items": [
    {
      "id": "req-open-1205",
      "timestamp": "2026-05-11T20:32:10+07:00",
      "room_id": "1205",
      "action": "open_room",
      "stay_mode": "overnight",
      "case": "overnight",
      "actor": "roomboy02",
      "result": "accepted",
      "detail": "guest_request"
    }
  ]
}
```

### `GET /api/alerts`

```json
{
  "items": [
    {
      "id": "alert-open-room-req-open-1205",
      "timestamp": "2026-05-11T20:35:00+07:00",
      "room_id": "1205",
      "severity": "high",
      "rule_code": "OPEN_ROOM_WHILE_VACANT",
      "status": "open",
      "title": "Open room while vacant",
      "detail": "Open-room event was recorded while room state remained vacant."
    }
  ]
}
```

---

## 6. SSE Contract

### `GET /events`

สถานะปัจจุบัน:

- frontend scaffold รองรับ SSE client แล้ว
- backend เปิด `GET /events` แล้ว
- เมื่อ client เชื่อมสำเร็จ backend จะส่ง `full_state` หนึ่งครั้งทันที
- หลังมี mutation ของห้อง backend จะส่ง `room_update` สำหรับห้องที่เปลี่ยน
- backend จะส่ง `audit_event` ทุกครั้งที่มี audit entry ใหม่
- backend จะส่ง `fraud_alert` เมื่อ rule ที่รองรับในปัจจุบันถูก trigger
- rule realtime ที่รองรับแล้วตอนนี้คือ `OPEN_ROOM_WHILE_VACANT`

Supported events:

- `full_state`
- `room_update`
- `audit_event`
- `fraud_alert`
- `gateway_status`
- `command_result`

ตัวอย่าง SSE:

```text
event: room_update
data: {"room_id":"1205","status":"cleaning","stay_mode":"overnight","case":"overnight","request_id":"01JV7R1H2CNH04H8J01KP6X1QV"}
```

```text
event: command_result
data: {"request_id":"01JV7R1H2CNH04H8J01KP6X1QV","room_id":"1205","result":"success"}
```

---

## 7. Error Contract

ทุก error response ควรใช้รูปแบบ:

```json
{
  "error": {
    "code": "COMMAND_TIMEOUT",
    "message": "Gateway did not return ACK in time",
    "request_id": "01JV7R1H2CNH04H8J01KP6X1QV"
  }
}
```

Suggested error codes:

- `UNAUTHORIZED`
- `ROOM_NOT_FOUND`
- `INVALID_ACTION`
- `COMMAND_TIMEOUT`
- `COMMAND_REJECTED`
- `CONTROLLER_OFFLINE`
- `PAYMENT_REQUIRED`

---

## 8. Database Model

backend ใหม่ควรเก็บ field ที่ต่อยอดมาจากคู่มือ master เดิมดังนี้

### `rooms`

```sql
CREATE TABLE rooms (
    room_id              TEXT PRIMARY KEY,
    floor                INTEGER NOT NULL,
    controller_id        TEXT NOT NULL,
    rs485_address        INTEGER NOT NULL,
    card_index           INTEGER NOT NULL,
    port_index           INTEGER NOT NULL,
    display_enabled      INTEGER NOT NULL DEFAULT 1
);
```

### `stays`

```sql
CREATE TABLE stays (
    stay_id              TEXT PRIMARY KEY,
    room_id              TEXT NOT NULL,
    stay_mode            TEXT NOT NULL,
    case_code            TEXT NOT NULL,
    status               TEXT NOT NULL,
    guest_in_at          TEXT,
    guest_out_at         TEXT,
    clean_up_at          TEXT,
    used_minutes         INTEGER,
    payment_amount       REAL,
    payment_ref          TEXT,
    remark               TEXT
);
```

### `audit_events`

```sql
CREATE TABLE audit_events (
    event_id             TEXT PRIMARY KEY,
    request_id           TEXT,
    room_id              TEXT NOT NULL,
    action               TEXT NOT NULL,
    stay_mode            TEXT,
    case_code            TEXT,
    actor_user_id        TEXT,
    actor_device_id      TEXT,
    result               TEXT NOT NULL,
    event_time           TEXT NOT NULL,
    remark               TEXT
);
```

DB rule ที่แนะนำ:

- `stay_mode` คือ canonical field ของระบบใหม่
- `case_code` ใช้เก็บค่าที่เข้ากับ report เดิม
- ถ้ายังไม่มี business differentiation มาก ให้ตั้ง `case_code = stay_mode`
