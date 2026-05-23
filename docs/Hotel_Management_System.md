# Hotel Management System
## Complete Project Blueprint

อ่านไฟล์นี้ไฟล์เดียว แล้วสร้างโปรเจคในรูปแบบนี้ได้ทันที

---

# 1. System Overview

ระบบ Hotel Maid Service Gateway ทำงานบน embedded board (Luckfox, Raspberry Pi) เป็น local IoT gateway เชื่อม hardware RS485/HC-12 กับ dashboard และ handheld app ผ่าน MQTT และ SSE

**Core philosophy:**
- Backend คือ source of truth ของ application state ทั้งหมด
- State ทั้งหมดอยู่ใน RAM — SQLite เก็บเฉพาะ config
- Event-driven: hardware reports → backend decides → clients display
- Thin client: dashboard/app แค่แสดง state และส่ง command
- Low disk write: ป้องกัน eMMC/SD เสื่อม — log ลง tmpfs เท่านั้น
- ห้าม update state ก่อน hardware ACK

---

# 2. Hardware Topology

```
PIC Devices (RS485 / HC-12, Protocol V3)
        ↓ serial port ต่อตรงกับ board
[Optional] ESP32-P4 Floor Gateway (MQTT → local broker)
        ↓
Board: Pi1 / Luckfox Lyra / Nova / Pi Zero W / Linaro
 ├── Mosquitto (local MQTT broker)
 ├── Go Backend (:8080)
 │    ├── RAM State Store
 │    ├── Protocol V3 Runtime (serial master)
 │    ├── MQTT Bridge (external broker)
 │    ├── REST API + SSE (/api, /events)
 │    ├── Flutter dashboard static files (optional, same origin)
 │    ├── Offline Detection
 │    └── SQLite (config only)
 ├── Nginx (optional static reverse proxy)
 └── frpc (remote access tunnel, public port 6010)
        ↓
External MQTT Broker (34.87.151.202:1883)
        ↑↓
Flutter Dashboard (web) / Flutter Handheld (PWA + mobile)
```

**Serial port per board:**

| Board | Port | หมายเหตุ |
|---|---|---|
| Luckfox Lyra | `/dev/ttyS0` | RS485 UART0 |
| Luckfox Nova | `/dev/ttyS4` | HC-12 UART4 (ต้องใช้ patched boot image) |
| Raspberry Pi 1 | `/dev/serial0` หรือ `/dev/ttyUSB0` | |
| Raspberry Pi Zero W | `/dev/serial0` หรือ `/dev/ttyUSB0` | |
| Linaro | `/dev/ttyUSB_RS485` | udev rule ผูกชื่อถาวร |

---

# 3. Repository Structure

```
project/
├── go-backend/               ← Go source (single binary)
│   ├── main.go               ← entry point, wire services
│   ├── app.go                ← App struct, OnRoomUpdate, applyRoomAction
│   ├── config.go             ← AppConfig จาก .env
│   ├── protocol_v3_runtime.go← serial master, scan loop
│   ├── room_state.go         ← RoomStateStore (RAM cache)
│   ├── room_db.go            ← SQLite CRUD
│   ├── bridge_service.go     ← MQTT bridge
│   ├── router.go             ← REST + SSE route command → applyRoomAction
│   └── push_notifications.go ← Web Push VAPID
├── flutter-dashboard/        ← supervisor UI (Flutter web, talks to REST + SSE)
├── flutter-handheld/         ← maid app (Flutter PWA + mobile)
├── flutter-mqtt/             ← MQTT monitor/admin (Flutter web)
├── pic-firmware/             ← PIC firmware (MPLAB X)
├── esp32-gateway/            ← optional ESP32-P4 floor gateway (ESP-IDF)
├── boards/
│   ├── lyra/                 ← docs, scripts, SDK, artifacts เฉพาะ Lyra
│   ├── nova/                 ← docs, scripts, SDK, artifacts เฉพาะ Nova
│   ├── raspi1/               ← docs, scripts เฉพาะ Pi1
│   ├── raspi0w/              ← docs, scripts เฉพาะ Pi Zero W
│   └── linaro/               ← docs, scripts เฉพาะ Linaro
├── dist/
│   ├── lyra/                 ← binary สำหรับ Lyra (armv7l)
│   ├── nova/                 ← binary สำหรับ Nova (arm64)
│   ├── raspi1/               ← binary สำหรับ Pi1 (ARMv6)
│   ├── raspi0w/              ← binary สำหรับ Pi Zero W (ARMv6)
│   └── linaro/               ← binary สำหรับ Linaro (arm64)
├── scripts/
│   ├── build.sh              ← cross-compile + flutter build
│   ├── deploy.sh             ← push ลงบอร์ด (ADB/SSH)
│   └── esp32-gateway.sh      ← build/flash ESP32 (wraps idf.py)
└── docs/                     ← protocol, state machine, API contract
```

---

# 4. Protocol V3 Hotel Edition

## 4.1 Frame Structure

```
AA  NFC      LEN       ADDR  COUNT  DATA      CRC_LO  CRC_HI
    [FL][FC] [DIR][L]
```

| Field | Size | Description |
|---|---|---|
| AA | 1 byte | Room/Node Address (0x00 = broadcast) |
| NFC | 1 byte | bit7-4=FLOOR, bit3-0=FC |
| LEN | 1 byte | bit7=DIR, bit6-0=Payload Length |
| ADDR | 1 byte | Register address |
| COUNT | 1 byte | Number of data bytes |
| DATA | variable | payload |
| CRC_LO/HI | 2 bytes | CRC16 Modbus (poly=0xA001, init=0xFFFF) |

Payload Length = ADDR(1) + COUNT(1) + DATA(n)

## 4.2 Function Codes

| FC | Name | Description |
|---|---|---|
| 0x01 | Read | Master อ่าน register จาก node |
| 0x02 | Write | Master เขียน register ไป node |
| 0x03 | Ping | Ping node |
| 0x04 | Device Info | ขอข้อมูล device |
| 0x05 | Event Push | Node push state มาหา master (DIR=1, unsolicited) |

## 4.3 FLOOR Values

| Value | Meaning |
|---|---|
| 0x0 | All Floors (global broadcast) |
| 0x1-0xF | Floor 1-15 |

## 4.4 Frame Examples

**Read Inpbuf[0] of Room 5 Floor 2:**
```
05 21 02 00 01 CRC_LO CRC_HI
```

**Response:**
```
05 21 83 00 01 28 CRC_LO CRC_HI
```
(LEN=0x83: DIR=1, payload=3)

**Write Outbuf[4] of Room 5 Floor 2 (2 bytes, big-endian):**
```
05 22 06 14 02 00 03 CRC_LO CRC_HI
```

**Write ACK:**
```
05 22 82 14 02 CRC_LO CRC_HI
```
(LEN=0x82: DIR=1, payload=2, no data)

**Error Response:**
```
AA NFC LEN ADDR 00 ERROR_CODE CRC_LO CRC_HI
```
COUNT=0x00, DATA[0]=error code

**Broadcast write DND to all rooms on Floor 3:**
```
00 32 03 14 01 01 CRC_LO CRC_HI
```

**Event Push — Inpbuf[0]:**
```
05 25 83 00 01 28 CRC_LO CRC_HI
```
(DIR=1, ADDR=0x00, COUNT=1)

**Event Push — Outbuf[4]:**
```
05 25 84 14 02 00 03 CRC_LO CRC_HI
```
(DIR=1, ADDR=0x14, COUNT=2, big-endian word)

## 4.5 Error Codes

| Code | Meaning |
|---|---|
| 0x01 | Invalid Function |
| 0x02 | Invalid Address |
| 0x03 | Invalid Count |
| 0x04 | Busy / Not Ready |
| 0x05 | Write Protected |
| 0x06 | CRC Error |
| 0x07 | Internal Error |

## 4.6 Protocol Timing

| Parameter | Value |
|---|---|
| Timeout per ACK | 600 ms |
| Post-write delay | 20 ms |
| Retry interval | 620 ms |
| Max retries | 5 |

**Full scan strategy:**
1. อ่าน `Inpbuf[0]` ก่อน
2. ถ้าไม่ ACK → mark offline → ข้ามห้องนั้น
3. ถ้า ACK → อ่าน `Outbuf[4]`

วิธีนี้ลด frame ที่ไม่จำเป็นและทำให้ link HC-12 นิ่ง

## 4.7 CRC16 Modbus

```c
uint16_t crc16Modbus(uint8_t *data, size_t len) {
    uint16_t crc = 0xFFFF;
    for (size_t i = 0; i < len; i++) {
        crc ^= data[i];
        for (int b = 0; b < 8; b++) {
            if (crc & 0x0001) crc = (crc >> 1) ^ 0xA001;
            else crc >>= 1;
        }
    }
    return crc;  // transmit: CRC_LO first, CRC_HI second
}
```

---

# 5. Register Map

## Input Registers

### 0x00 — REG_INP0_ROOM_STATUS (Inpbuf[0])

```c
#define Dndin   Bit0   // DND Input
#define Murin   Bit1   // Makeup Room Input
#define Bellin  Bit2   // Bell Input
#define Guein   Bit3   // Guest Input     ← ใช้งานหลัก
#define Maidin  Bit4   // Maid Input      ← ใช้งานหลัก
#define Ccsin   Bit5   // CCS/PMS Input
#define Masin   Bit6   // Master Input
#define Doorin  Bit7   // Door Sensor Input
```

## Output Registers

### 0x10-0x13 — REG_OUT0-3_LAMP (Outbuf[0-3])
Output relay bits: Out1-Out8 per register (Bit0-Bit7)

```c
// Outbuf[0] lamp relay aliases
#define Lamp1   Out1  // = Bit0
#define Lamp2   Out2  // = Bit1
#define Lamp3   Out3  // = Bit2
#define Lamp4   Out4  // = Bit3
```

### 0x14 — REG_OUT4_CONTROL (Outbuf[4]) ← ใช้งานหลัก

```c
#define Dnd      Bit0   // DND Output
#define Mur      Bit1   // Makeup Room Output
#define Aux1Out  Bit2   // UrgentMUR signal ใน hotel edition
#define Aux2Out  Bit3
#define Door     Bit4   // Door Output
#define Ccs      Bit5   // CCS/PMS Output
#define Maid     Bit6   // Maid Output
#define Guest    Bit7   // Guest Output
#define UrgMur   Aux1Out  // alias
```

### 0x15 — REG_AIR_CONTROL (Outbuf[5])

```c
#define Temp     0x0F   // Temperature mask (TempValue = (reg & 0x0F) + 15)
#define FanSpeed 0x30   // Fan speed mask
#define AirCdu   Bit6   // Condenser enable
#define AirPower Bit7   // AC power
#define FanL  0x00  #define FanM  0x10  #define FanH  0x20  #define FanA  0x30
```

### 0x16-0x1B — Time / Alarm / Node ID

| Address | Register |
|---|---|
| 0x16 | REG_TEMP_ACTUAL |
| 0x17 | REG_REAL_HOUR |
| 0x18 | REG_REAL_MIN |
| 0x19 | REG_ALARM_HOUR |
| 0x1A | REG_ALARM_MIN |
| 0x1B | REG_NODE_ID |

**Backend ใช้จริงเฉพาะ 2 register:**
- `Inpbuf[0]` ที่ address `0x00`
- `Outbuf[4]` ที่ address `0x14` (read/write 2 bytes, big-endian word)

---

# 6. Room Status & State Machine

## 6.1 Status Values

| Status | Hex Color | ความหมาย | Hardware condition |
|---|---|---|---|
| `vacant` | `#9CA3AF` | ห้องว่าง | Guest=0, Maid=0, DND=0, MUR=0 |
| `occupied` | `#22C55E` | มีแขก | Guein=1 หรือ Outbuf[4] Guest=1 |
| `dnd` | `#EF4444` | ห้ามรบกวน | Outbuf[4] DND=1 |
| `makeup` | `#F59E0B` | รอทำห้อง | Outbuf[4] MUR=1 |
| `urgent_makeup` | `#F97316` | ด่วน | Outbuf[4] Bit2=1 (UrgMur) |
| `cleaning` | `#3B82F6` | กำลังทำ | Maidin=1 หรือ cleaning command |
| `offline` | `#374151` | ไม่ตอบสนอง | full scan ไม่ได้ ACK |

## 6.2 Status Decode Logic

```go
// Priority: outbuf4 > inpbuf0
func DecodeEffectiveStatus(inpbuf0, outbuf4 *int) string {
    if outbuf4 != nil && hasOutbufStatusBits(*outbuf4) {
        return DecodeStatusFromOutbuf(*outbuf4)
    }
    if inpbuf0 != nil {
        return DecodeStatus(*inpbuf0)
    }
    return "vacant"
}

func DecodeStatusFromOutbuf(outbuf4 int) string {
    if outbuf4 & 0x0001 != 0 { return "dnd" }        // DND bit0
    if outbuf4 & 0x0040 != 0 { return "cleaning" }    // Maid bit6
    if outbuf4 & 0x0004 != 0 { return "urgent_makeup" }// UrgMur bit2
    if outbuf4 & 0x0002 != 0 { return "makeup" }      // MUR bit1
    if outbuf4 & 0x0080 != 0 { return "occupied" }    // Guest bit7
    return "vacant"
}

func DecodeStatus(inpbuf0 int) string {
    if inpbuf0 & 0x0010 != 0 { return "cleaning" }    // Maidin bit4
    if inpbuf0 & 0x0008 != 0 { return "occupied" }    // Guein bit3
    return "vacant"
}
```

## 6.3 State Transition Flow

**Guest Flow:**
```
vacant → (card in) → occupied → (press DND) → dnd
       → (pull card + 30s timeout) → vacant
       → (press MUR) → makeup → (maid accept) → cleaning
                              → (maid finish) → vacant
```

**Business logic interlocks:**
- DND set → clear MUR, clear UrgMur
- MUR set → clear DND
- Maid mode → lock AC (energy saving)
- Backend cleaning state persists ถ้า hardware MUR bit ยังค้าง (รอ finish_cleaning command)

## 6.4 Command → Register Patch

| Action | Outbuf[4] change |
|---|---|
| `makeup_room` | set MUR (bit1), clear DND (bit0), clear UrgMur (bit2) |
| `urgent_cleaning` | set UrgMur (bit2), clear MUR (bit1), clear DND (bit0) |
| `start_cleaning` | preserve ค่าเดิม, set MUR |
| `finish_cleaning` | clear MUR (bit1), clear UrgMur (bit2) |

**กฎสำคัญ: backend เขียน register ไป PIC → รอ ACK → อัปเดต state → publish**
ถ้า PIC ไม่ ACK ห้าม update state และห้าม publish

---

# 7. Backend Architecture (Go)

## 7.1 Service Wiring (main.go)

```go
cfg := LoadConfig()           // .env
store := NewRoomStateStore()  // RAM state
runtime := NewProtocolRuntime(cfg)  // serial master
bridge := NewBridgeService(cfg)     // MQTT
api := NewHTTPServer(cfg, store, runtime, bridge) // REST + SSE + dashboard static
app := NewApp(cfg, store, runtime, bridge, api)

go runtime.Start(ctx, store, onUpdate, onLinkStatus)
go bridge.Start(ctx)
go api.Start(ctx)
app.RunEventLoop(ctx)
```

## 7.2 Key Data Flow

```
RS485/serial poll → ProtocolRuntime → store.Update(deviceID, inpbuf0, outbuf4)
                                            → App.OnRoomUpdate(room)
                                                  → MQTT publish hotel/dashboard/update
                                                  → SSE push room_update
                                                  → PushManager.NotifyWithCooldown()

Client command → applyRoomAction(action, room)
                      → runtime.WriteRegister(deviceID, 0x14, newOutbuf4)
                            → wait ACK
                            → store.UpdateFromStatus(roomNo, status, ...)
                            → App.OnRoomUpdate(room)
```

## 7.3 State Update Rule

```go
// Patch เฉพาะ bit ที่เกี่ยวข้อง ไม่ overwrite ทั้ง register
func mergeMaskedBits(current, update, mask int) int {
    return (current &^ mask) | (update & mask)
}

// ห้าม write 0x0000 ไป register
// ห้าม update store ก่อน hardware ACK
```

## 7.4 Offline Detection

```go
// เก็บ last_seen per device
// ถ้า now - last_seen > threshold → SetOffline(deviceID)
// threshold แนะนำ 30-60 วินาที
// full resync ทุก 10-15 นาที เพื่อ recover missed event
```

## 7.5 SQLite Schema (config only)

```sql
-- rooms: static config เท่านั้น
CREATE TABLE rooms (
    device_id   INTEGER PRIMARY KEY,
    room        INTEGER,
    floor       INTEGER,
    proto_room  INTEGER,
    maid_name   TEXT
);

-- maid assignments
CREATE TABLE maid_assignments (
    device_id   INTEGER,
    maid_name   TEXT,
    is_primary  INTEGER DEFAULT 0
);
```

ห้ามเก็บ realtime state, log, หรือ polling history ใน SQLite

## 7.6 Config (.env)

```env
SERIAL_PORT=/dev/ttyS4
MOCK_MODE=false
BAUD_RATE=9600
BYTE_SIZE=8
PARITY=N
STOP_BITS=1
PROTOCOL_TIMEOUT=0.6
POST_WRITE_DELAY=0.02
PROTOCOL_RETRIES=5
EXTERNAL_BROKER_HOST=34.87.151.202
EXTERNAL_BROKER_PORT=1883
EXTERNAL_USERNAME=...
EXTERNAL_PASSWORD=...
PUSH_STORAGE_PATH=/tmp/hotel-maid-backend/push_subscriptions.json
PROTOCOL_TRACE=false
```

---

# 8. MQTT Topics

## 8.1 Overview

| Topic | Direction | QoS | Description |
|---|---|---|---|
| `hotel/dashboard/update` | Backend → Client | 1 | full_state (retained) / room_update |
| `hotel/dashboard/protocol_link` | Backend → Client | 0 | RS485 link status |
| `hotel/dashboard/command` | Client → Backend | 0 | room commands |
| `hotel/request/full_state` | Client → Backend | 0 | request snapshot |
| `hotel/admin/state` | Backend → Client | 1 | room map (retained) |
| `hotel/admin/rooms/set_all` | Client → Backend | 0 | replace entire room map |
| `hotel/admin/floors/assign` | Client → Backend | 0 | assign maids to floor |
| `hotel/admin/rooms/ack` | Backend → Client | 1 | command result |
| `hotel/admin/floors/ack` | Backend → Client | 1 | command result |

## 8.2 Payloads

**full_state:**
```json
{
  "type": "full_state",
  "rooms": [
    { "floor": 1, "room": 201, "guest": "off", "maid": "off",
      "maids": ["ดาว"], "status": "vacant", "time": "14:01" }
  ]
}
```

**room_update:**
```json
{
  "type": "room_update",
  "floor": 1, "room": 201,
  "guest": "off", "maid": "on",
  "maids": ["ดาว"], "status": "cleaning", "time": "14:05"
}
```

**room command:**
```json
{
  "type": "room_command",
  "floor": 1, "room": 201,
  "action": "start_cleaning",
  "time": "14:30"
}
```
actions: `makeup_room`, `urgent_cleaning`, `start_cleaning`, `finish_cleaning`

**set_all:**
```json
{
  "rooms": [
    { "device_id": 1, "room": 201, "floor": 1, "proto_room": 1, "maids": ["ดาว"] }
  ]
}
```

**floors/assign:**
```json
{ "floor": 2, "maids": ["ฝน", "ดาว"] }
```

## 8.3 Floor Gateway Topics (local broker)

```
hotel/gateway/{id}/hello        Gateway → Backend  (announce + capabilities)
hotel/gateway/{id}/heartbeat    Gateway → Backend  (health, last_seen)
hotel/gateway/{id}/snapshot     Gateway → Backend  (full floor state on reconnect)
hotel/gateway/{id}/event        Gateway → Backend  (single room state change)
hotel/gateway/{id}/command_ack  Gateway → Backend  (command result)
hotel/gateway/{id}/command      Backend → Gateway  (write register command)
```

---

# 9. REST + SSE API

```
GET  /api/state                    → full_state JSON
GET  /api/admin/state              → admin_state JSON
GET  /events                       → SSE stream (room_update, full_state, protocol_link)
POST /api/rooms/{room}/{action}    → room action
POST /api/rooms/{room}/command     → { "action": "..." }
POST /api/floors/{floor}/assign    → { "maids": [...] }
POST /api/push/subscribe           → push subscription
POST /api/push/unsubscribe
POST /api/push/test
GET  /ping                         → health check
```

SSE events ใช้ event shape กลางของระบบ: `full_state`, `room_update`, `audit_event`, `fraud_alert`, `command_result`, `protocol_link`, `push_config`

---

# 10. Build & Deploy

## 10.1 Cross-compile Matrix

| BOARD | GOOS | GOARCH | GOARM | Deploy |
|---|---|---|---|---|
| lyra | linux | arm | 7 | ADB |
| nova | linux | arm64 | — | ADB (`-s 89bc70dc1ea69f42`) |
| raspi1 | linux | arm | 6 | SSH |
| raspi0w | linux | arm | 6 | SSH |
| linaro | linux | arm64 | — | SSH |

```bash
# Go cross-compile ตัวอย่าง (Nova)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../dist/nova/hotel-maid-backend .

# Flutter web
flutter build web --dart-define-from-file=config/dart_defines.web.json --release
```

## 10.2 Build Commands

```bash
./scripts/build.sh all                    # Lyra
BOARD=nova ./scripts/build.sh all         # Nova
BOARD=raspi1 ./scripts/build.sh all       # Pi1
BOARD=raspi0w ./scripts/build.sh all      # Pi Zero W
```

## 10.3 Deploy Commands

```bash
# ADB boards
./scripts/deploy.sh all
BOARD=nova ./scripts/deploy.sh all
adb shell '/etc/init.d/S95hotel-maid-backend restart'

# SSH boards
BOARD=raspi1 PI_HOST=192.168.0.204 PI_USER=kaina ./scripts/deploy.sh all
ssh kaina@192.168.0.204 'sudo systemctl restart hotel-maid-backend nginx'
```

## 10.4 ESP32 Gateway Build

```bash
# ต้อง source IDF ก่อน (ESP-IDF v5.3+ required for ESP32-P4)
. ~/esp/esp-idf/export.sh
./scripts/esp32-gateway.sh build
./scripts/esp32-gateway.sh flash-monitor

# Override port
ESPPORT=/dev/ttyACM0 ./scripts/esp32-gateway.sh flash-monitor
```

Script defaults: `IDF_TARGET=esp32p4`, `ESPPORT=/dev/ttyACM0`, `ESPBAUD=460800`

---

# 11. Remote Access (frpc)

```
Internet → frps (VPS: 34.87.151.202:6010) → frpc (on board) → Go Backend (:6010) → Dashboard + REST/SSE
```

**Port plan:**

| Board | Dashboard | SSH |
|---|---|---|
| Nova | 6001 | 6101 |
| Lyra | 6010 | 6110 |
| Pi1 | 6010 | 6110 |
| Linaro | 6004 | 6104 |
| Pi Zero W | 6005 | 6105 |
| CCS-2PLUS dev | 6010 | 6110 |

ใช้ frpc เท่านั้น — ไม่ใช้ Tailscale หรือ VPN

สำหรับ CCS-2PLUS ปัจจุบัน ให้ใช้ public endpoint เดียว:

```text
http://34.87.151.202:6010/
```

เส้นทางสำคัญบน endpoint เดียวกัน:

```text
/              → Flutter dashboard static
/api/state     → REST API
/events        → SSE realtime stream
```

บน `raspi1` ที่ใช้งานจริง:

- `hotel-maid-backend` รันผ่าน `systemd`
- `frpc` รันผ่าน `systemd`
- `frpc` config หลักคือ `frpc/ccs2plus-6010.toml`
- public dashboard/backend ใช้ `6010`
- public SSH ใช้ `6110`

---

# 12. Storage Policy

| Data | Location | Rule |
|---|---|---|
| Go binary, `.env` | eMMC | เขียนตอน deploy เท่านั้น |
| `hotel.db` (SQLite) | eMMC | เขียนตอน config เปลี่ยนเท่านั้น |
| `vapid_keys.json` | eMMC | generate ครั้งเดียว |
| Runtime logs | `/run/log/journal` (tmpfs via `journald` volatile) | ไม่เขียน eMMC/SD |
| PID file | `/var/run/` (tmpfs) | |
| Push subscriptions | `/tmp/` (tmpfs) | reset ได้ตาม reboot |
| Realtime state | RAM only | ไม่มี disk write ใดๆ |

**ห้ามเก็บ:**
- realtime log ลง disk
- polling history
- current state ลง SQLite

---

# 13. Flutter Architecture

## 13.1 Config Injection

```bash
# Web/PWA — inject via dart-define
flutter build web --dart-define-from-file=config/dart_defines.web.json --release

# Mobile
flutter build apk --dart-define-from-file=config/dart_defines.mobile.json --release
```

## 13.2 Key dart-defines

```json
{
  "BACKEND_BASE_URL": "http://34.87.151.202:6010",
  "MQTT_HOST": "34.87.151.202",
  "MQTT_PORT": "1883"
}
```

## 13.3 REST + SSE Flow (dashboard)

```
browser → same-origin /api/* and /events → Go backend
```

เมื่อ backend เสิร์ฟ Flutter dashboard static เอง สามารถ build dashboard ด้วย same-origin config ได้:

```bash
flutter build web --dart-define=BACKEND_BASE_URL= --release
```

ในโหมด same-origin:
- REST base URL คือ origin ปัจจุบัน เช่น `http://34.87.151.202:6010`
- SSE URL คือ `${BACKEND_BASE_URL}/events` หรือ `/events` เมื่อ `BACKEND_BASE_URL` ว่าง

## 13.4 Status Color Palette

```dart
// lib/theme/status_palette.dart
vacant:        Color(0xFF9CA3AF)
occupied:      Color(0xFF22C55E)
dnd:           Color(0xFFEF4444)
makeup:        Color(0xFFF59E0B)
urgent_makeup: Color(0xFFF97316)
cleaning:      Color(0xFF3B82F6)
offline:       Color(0xFF374151)
```

## 13.5 Dashboard Summary Groups

| Group | Statuses |
|---|---|
| READY | vacant |
| NEW | makeup, urgent_makeup |
| DOING | cleaning |
| DND | dnd |
| OVERDUE | urgent_makeup ที่เกินเวลา |

---

# 14. ESP32-P4 Floor Gateway

## 14.1 Role

Optional floor gateway: poll PIC nodes ผ่าน HC-12/RS485 แล้ว publish ผ่าน MQTT ไป local broker บน backend board

Backend ยังคงเป็น source of truth — ESP32 เป็นแค่ transport layer

## 14.2 Firmware Structure

```
esp32-gateway/
├── CMakeLists.txt
├── sdkconfig.defaults        ← CONFIG_IDF_TARGET="esp32p4"
└── main/
    ├── CMakeLists.txt        ← REQUIRES: mqtt, json; PRIV_REQUIRES: nvs_flash
    ├── app_main.c
    ├── gateway_app.c         ← MQTT lifecycle, reconnect, heartbeat loop
    ├── gateway_protocol.c    ← JSON message formatters/parsers
    ├── gateway_protocol.h
    └── gateway_config.h      ← GATEWAY_ID, GATEWAY_FLOOR, BROKER_HOST, ROOM_COUNT
```

## 14.3 gateway_config.h Template

```c
#define GATEWAY_PROTOCOL_VERSION 1
#define GATEWAY_FW_VERSION "0.1.0"
#define GATEWAY_MODEL "esp32-p4-nano"
#define GATEWAY_PROTOCOL_MODE "v3-hotel-edition-floor-gateway"
#define GATEWAY_ID "floor-2-p4-01"
#define GATEWAY_FLOOR 2
#define GATEWAY_BROKER_HOST "192.168.0.10"
#define GATEWAY_BROKER_PORT 1883
#define GATEWAY_TOPIC_ROOT "hotel/gateway"
#define GATEWAY_HEARTBEAT_SEC 10
#define GATEWAY_RECONNECT_DELAY_MS 3000
#define GATEWAY_RX_BUFFER_SIZE 1024
#define GATEWAY_TX_BUFFER_SIZE 1024
#define GATEWAY_ROOM_COUNT 3
```

## 14.4 Message Formats (JSON over MQTT)

**hello:**
```json
{
  "v": 1, "type": "hello", "gateway_id": "floor-2-p4-01", "seq": 1,
  "body": {
    "floor": 2, "model": "esp32-p4-nano", "fw_version": "0.1.0",
    "rooms": [201, 202, 203],
    "capabilities": ["snapshot", "event", "command_ack", "heartbeat"]
  }
}
```

**event (single room change):**
```json
{
  "v": 1, "type": "event", "gateway_id": "floor-2-p4-01",
  "body": {
    "floor": 2, "room": 201, "proto_room": 1,
    "reason": "inpbuf0_change",
    "inpbuf0": 8, "outbuf4": 3, "link": "online"
  }
}
```

**command (backend → gateway):**
```json
{
  "command_id": "abc123",
  "room": 201, "proto_room": 1,
  "target": { "outbuf4": 3 }
}
```

---

# 15. PIC Firmware Structure

```c
// main.c — key sections

/* Input bits (Inpbuf[0] / 0x00) */
#define Guein   Bit3    // Guest Input
#define Maidin  Bit4    // Maid Input
#define Doorin  Bit7    // Door Sensor

/* Output control (Outbuf[4] / 0x14) */
#define Dnd     Bit0
#define Mur     Bit1
#define UrgMur  Bit2    // = Aux1Out
#define Door    Bit4
#define Maid    Bit6
#define Guest   Bit7

/* Lamp relay aliases (Outbuf[0] / 0x10) */
#define Lamp1   Out1    // Bit0
#define Lamp2   Out2    // Bit1
#define Lamp3   Out3    // Bit2
#define Lamp4   Out4    // Bit3

/* Data types */
uint8_t  Inpbuf[6];     // input registers
uint16_t Outbuf[12];    // output registers (use cast to avoid warning 373)

// warning 373 fix: cast ทุกครั้งที่ assign int expression ให้ uint16_t
Outbuf[0] = (uint16_t)(Lamp1 | Lamp2 | Lamp3 | Lamp4);
Outbuf[4] = (uint16_t)(value);
```

**Protocol constants:**
```c
#define PROTOCOL_FC_READ    0x01
#define PROTOCOL_FC_WRITE   0x02
#define PROTOCOL_FC_EVENT   0x05
#define PROTOCOL_REG_INPBUF0  0x00
#define PROTOCOL_REG_OUTBUF4  0x14
```

---

# 16. Design Rules (Non-negotiable)

1. **Hardware ACK before state update** — ห้าม update store หรือ publish ก่อนได้ ACK จาก PIC
2. **Patch bits only** — ห้าม overwrite ทั้ง register ด้วย 0x0000, patch เฉพาะ bit ที่เกี่ยวข้อง
3. **RAM-first** — realtime state ทั้งหมดอยู่ใน RAM ไม่มี disk write ในช่วง runtime ปกติ
4. **No Docker** — single binary, systemd/init.d, ไม่ใช้ container บน embedded board
5. **BOARD env pattern** — code base เดียว, `BOARD` เป็น build-time flag เท่านั้น
6. **No dart hardcode** — URL, host, port inject ผ่าน `--dart-define` ทั้งหมด
7. **frpc only** — remote access ผ่าน frpc เท่านั้น ไม่ใช้ VPN
8. **REST + SSE only** — dashboard และ handheld ใช้ REST สำหรับ command/query และ SSE สำหรับ realtime
9. **MQTT as transport** — ไม่ใช้ MQTT เป็น database หรือ state store
10. **One source of truth** — backend Go เท่านั้น ห้าม client ถือ business logic

---

# 17. New Project Checklist

เมื่อสร้างโปรเจคใหม่ในรูปแบบนี้:

- [ ] กำหนด `GATEWAY_ID`, `GATEWAY_FLOOR`, `GATEWAY_ROOM_COUNT` ใน `gateway_config.h`
- [ ] ตั้ง serial port ใน `.env` ให้ตรงกับบอร์ด
- [ ] สร้าง room map ใน SQLite ผ่าน `hotel/admin/rooms/set_all`
- [x] ใช้ `frpc` public port `6010` และ SSH port `6110`
- [ ] build Flutter dashboard ด้วย `BACKEND_BASE_URL` ที่ตรงกับ frpc endpoint หรือเว้นว่างเมื่อเสิร์ฟ same-origin จาก backend
- [ ] ตั้ง `PUSH_STORAGE_PATH` ไปที่ `/tmp/` บน eMMC boards
- [ ] ตั้ง `tmpfs` สำหรับ `/var/log/` และ `/var/run/` ใน fstab
- [ ] เปิด `MOCK_MODE=true` ตอน dev ก่อนต่อ hardware จริง
- [ ] ใช้ `PROTOCOL_TRACE=true` ตอน debug serial frame
- [ ] build ด้วย `CGO_ENABLED=0` เสมอ (no cgo on embedded)
- [ ] Flutter web ใช้ `--dart-define-from-file` เสมอ ห้าม hardcode URL
