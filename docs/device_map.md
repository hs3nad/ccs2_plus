# CCS-2PLUS PRO Device Map

เอกสารนี้สรุป `current software-only topology` ของระบบในเฟสปัจจุบัน

หมายเหตุ:

- ในเฟสนี้ `ยังไม่รวม PIC`
- ในเฟสนี้ `ยังไม่รวม ESP32 Gateway`
- ระบบที่ใช้งานตอนนี้เน้น `backend + dashboard + handheld + cashier + frpc`
- เอกสาร legacy integration ยังคงเก็บไว้ใน [docs/protocol.md](/home/kaina/work/project/ccs2_plus/docs/protocol.md) และ [docs/rs485_transport_spec.md](/home/kaina/work/project/ccs2_plus/docs/rs485_transport_spec.md) สำหรับใช้ภายหลัง

---

## 1. Device Layers

| Layer | Device Type | Role |
|---|---|---|
| Application | Flutter Dashboard | Front office + manager view |
| Application | Flutter Handheld | Mobile workflow สำหรับพนักงาน |
| Application | Flutter Cashier | Cashier / front office workflow |
| Backend | Go Backend | API, rules, audit, realtime state |
| Access | frpc | Remote access tunnel |

---

## 2. Current Deployment

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

กฎของ deployment นี้:

- `go-backend` เป็น source of truth
- client ทุกตัวคุยกับ backend ผ่าน `REST + SSE`
- deployment ปัจจุบันเปิด public endpoint ผ่าน `frpc` ที่ `http://34.87.151.202:6010/`
- remote SSH สำหรับ board ปัจจุบันใช้ `34.87.151.202:6110`
- runtime services บน `raspi1` ใช้ `systemd` สำหรับทั้ง `hotel-maid-backend` และ `frpc`
- `MQTT` ยังไม่ใช้
- `RS-485`, `PIC`, `ESP32 Gateway` ยังไม่อยู่ใน scope ของเฟสนี้

---

## 3. Identity Model

| ID | Example | Owner | Notes |
|---|---|---|---|
| `room_id` | `1205` | Backend | business room number |
| `floor_id` | `12` | Backend | ใช้ grouping |
| `device_id` | `HH-03` | Handheld/Cashier | physical client device |
| `user_id` | `roomboy02` | Backend auth | actor ของคำสั่ง |
| `stay_id` | `stay-01JV...` | Backend | stay/business record |
| `request_id` | `req-01JV...` | Backend/client | ใช้ trace คำสั่ง |

---

## 4. Room Data Model

ข้อมูลหลักที่ backend ควรถือ:

```json
{
  "room_id": "1205",
  "floor": 12,
  "status": "occupied",
  "stay_mode": "overnight",
  "case": "overnight",
  "remaining_minutes": 95
}
```

ความหมาย:

- `status` = สถานะ realtime ของห้อง
- `stay_mode` = รูปแบบการเข้าพัก
- `case` = field compatibility สำหรับ report legacy

---

## 5. Client Responsibilities

### Dashboard

- front office
- manager monitoring
- check-in / check-out / extend
- report / audit / alert view

### Handheld

- operation หน้างานแบบมือถือ
- open room
- start / finish cleaning
- task-oriented workflow

### Cashier

- front office workflow
- check-in / check-out usage
- room board and operator view

---

## 6. Deferred Integration

ส่วนที่ยังไม่ใช้ในเฟสนี้:

- PIC slave
- PIC display
- RS-485 transport
- ESP32 Gateway
- MQTT

เมื่อกลับมาเปิด scope นี้อีกครั้ง ให้ใช้อ้างอิงจาก:

- [docs/protocol.md](/home/kaina/work/project/ccs2_plus/docs/protocol.md)
- [docs/rs485_transport_spec.md](/home/kaina/work/project/ccs2_plus/docs/rs485_transport_spec.md)
