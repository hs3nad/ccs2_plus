# RS-485 Transport Service Spec

เอกสารนี้กำหนด service ฝั่ง `Raspberry Pi 1` สำหรับส่งคำสั่งจาก backend ไปยัง `PIC slave` และ `PIC display` ผ่าน `USB-RS485`

---

## 1. Purpose

service นี้รับผิดชอบ:

- รับ desired device state จาก backend
- resolve room mapping -> `rs485_address`, `card`, `port`
- encode เป็น `legacy slave frame` และ `legacy display frame`
- ส่ง frame ลง serial port
- จัดการ timeout, retry, sequencing และ error reporting

service นี้ไม่ถือ business logic หลักของโรงแรม

---

## 2. Default Deployment

```text
Handheld / Tablet
        |
        v
Go Backend on Raspberry Pi 1
        |
        +--> Room mapping store
        +--> RS-485 transport service
                  |
                  v
            USB-RS485 adapter
                  |
                  v
          PIC slave / PIC display
```

---

## 3. Responsibilities

### 3.1 Backend

- validate action
- decide target semantic state
- decide whether action changes relay state, display state, or both
- call transport service with resolved command intent

### 3.2 RS-485 Transport Service

- keep per-card last known payload image
- patch nibble for target room
- build slave checksum
- build display frame
- serialize writes to bus
- retry failed writes
- report success/failure back to backend

### 3.3 PIC Boards

- slave board applies output latch
- display board updates LED/blink state

---

## 4. Transport Inputs

transport service ควรรับ input ที่ผ่าน business decision มาแล้ว เช่น:

```json
{
  "request_id": "01JV...",
  "room_id": "1205",
  "relay_state": "cleaning",
  "display_state": "cleaning"
}
```

หรือในโค้ด:

```go
type CommandIntent struct {
    RequestID    string
    RoomID       string
    RelayState   RelayState
    DisplayState DisplayState
}
```

---

## 5. Room Mapping Requirements

backend ต้องมี mapping อย่างน้อย:

```json
{
  "room_id": "1205",
  "controller_id": "ctrl-1205",
  "rs485_address": 3,
  "card_index": 3,
  "port_index": 5
}
```

ข้อกำหนด:

- `rs485_address` ใช้ใน slave frame
- `card_index` และ `port_index` ใช้คำนวณ `PORT_REF` ของ display
- ถ้า `card_index` กับ `rs485_address` ต่างกัน ต้องเก็บแยกชัดเจน

---

## 6. Legacy Frame Rules

### 6.1 Slave Frame

```text
':' ADDR DATA_LO DATA_HI CHKSUM
```

กติกา:

- `ADDR` คือ address ของ card/slave
- `DATA_LO` และ `DATA_HI` คือ relay bitmask 16 บิตของ card นั้น
- `CHKSUM = (~sum) + 1`

### 6.2 Display Frame

```text
':' 0xFE PORT_REF LED_MODE
```

กติกา:

- `PORT_REF = (card_index * 16) + port_index`
- `LED_MODE`:
  - `0` off
  - `1` slow blink
  - `2` fast blink
  - `3` on

---

## 7. Semantic To Transport Mapping

| Semantic state | Slave status | Display mode |
|---|---|---|
| `vacant` | `1` | `0` |
| `occupied` | `2` | `3` |
| `cleaning` | `3` | `3` |
| `overstay` | `4` | `3` |

action mapping:

| Action | Relay target | Display target |
|---|---|---|
| `check_in` | `occupied` | `occupied` |
| `check_out` | `vacant` | `vacant` |
| `extend_stay` | `occupied` | `occupied` |
| `start_cleaning` | `cleaning` | `cleaning` |
| `finish_cleaning` | `vacant` | `vacant` |
| `set_urgent_makeup` | no relay change by default | `overstay` or site-specific alert blink |

---

## 8. Service API Suggestion

```go
type Service interface {
    ApplyIntent(ctx context.Context, intent CommandIntent) (CommandResult, error)
}
```

```go
type CommandResult struct {
    RequestID     string
    RoomID        string
    SlaveFrameHex string
    DisplayFrameHex string
    SentAt        time.Time
    RetryCount    int
}
```

---

## 9. Retry And Timeout

เพราะ legacy protocol ไม่มี rich ACK ชัดเจนในเอกสารปัจจุบัน ให้เริ่มต้นแบบ conservative:

- serialize writes ทีละ frame
- inter-frame delay สั้น ๆ เช่น `20-50ms`
- retry ได้ `1-3` ครั้ง
- ถ้าเกิน timeout ให้ mark `transport_failed`
- backend ควรแยก `command accepted` ออกจาก `transport delivered best-effort`

---

## 10. State Handling

transport service ควรถือข้อมูลชั่วคราวดังนี้:

- `last payload image` ต่อ `rs485_address`
- `last display mode` ต่อ `room_id`
- optional write queue

ข้อสำคัญ:

- ห้าม rebuild payload ของทั้ง card จากห้องเดียวแบบไม่รู้ค่าเดิม
- ต้อง patch เฉพาะ nibble ของ room เป้าหมาย
- initial payload image อาจเริ่มจากศูนย์หรือโหลดจาก persisted snapshot

---

## 11. Recommended Files For Pi1 Service

```text
go-backend/
  go.mod
  main.go
  internal/config/
  internal/roommap/
  internal/legacy/
  internal/transport/
```

---

## 12. Open Questions

- มี physical ACK จาก slave/display ที่ใช้งานได้จริงหรือไม่
- `open_room` ต้อง map เป็น output pulse แยกหรือไม่
- `D8` มีการใช้งานเฉพาะ site หรือไม่
- DND / makeup / urgent makeup มี LED/output เฉพาะจริงในหน้างานหรือไม่
