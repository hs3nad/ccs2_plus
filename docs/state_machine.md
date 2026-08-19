# CCS-2PLUS PRO State Machine

เอกสารนี้กำหนดสถานะห้องและ flow หลักของระบบ

---

## 1. Room Status

| Status | Meaning |
|---|---|
| `vacant` | ห้องว่าง |
| `occupied` | มีผู้เข้าพัก |
| `cleaning` | พนักงานกำลังทำความสะอาด |
| `dnd` | ห้ามรบกวน |
| `makeup` | รอทำความสะอาด |
| `urgent_makeup` | งานด่วน |
| `overstay` | เกินเวลาการเข้าพัก |
| `controller_offline` | gateway/PIC ไม่ตอบสนอง |
| `unknown` | state ยังไม่ชัดหรือรอ sync |

---

## 1.1 Stay Mode And Case

ระบบใหม่ควรแยก `room status` ออกจาก `stay_mode` และ `case`

| Field | Meaning | Example |
|---|---|---|
| `status` | สถานะ realtime ของห้อง | `occupied`, `cleaning` |
| `stay_mode` | รูปแบบการเข้าพัก | `overnight`, `temporary` |
| `case` | ค่า compatibility สำหรับ report เดิม | `overnight`, `temporary` |

recommended values:

- `overnight`
- `temporary`
- `monthly`
- `yearly`

หลักคิด:

- `status` เปลี่ยนได้บ่อยตาม runtime
- `stay_mode` เปลี่ยนตามชนิดการเช่าหรือชนิดการใช้งาน
- `case` เก็บไว้เพื่อรองรับคำศัพท์/format report แบบ CCS2 เดิม

---

## 2. Guest Lifecycle

```text
vacant
  -> check_in
occupied
  -> extend_stay
occupied
  -> timeout reached
overstay
  -> check_out
vacant
```

ตัวอย่าง:

- ห้องอาจมี `status=occupied`
- แต่มี `stay_mode=temporary`
- และ `case=temporary`

---

## 3. Housekeeping Lifecycle

```text
vacant
  -> set_makeup
makeup
  -> staff_accept / start_cleaning
cleaning
  -> finish_cleaning
vacant
```

กรณีเร่งด่วน:

```text
vacant or occupied
  -> set_urgent_makeup
urgent_makeup
  -> start_cleaning
cleaning
  -> finish_cleaning
vacant
```

---

## 4. DND Rules

```text
occupied
  -> set_dnd
dnd
  -> clear_dnd
occupied
```

กติกา:

- `dnd` ไม่ควรอยู่พร้อม `makeup`
- ถ้าตั้ง `makeup` ให้ clear `dnd`
- ถ้าตั้ง `dnd` ให้ clear `makeup` และ `urgent_makeup`

---

## 5. Command Commit Rule

```text
client command
  -> backend validate
  -> send to RS-485 transport
  -> transport send to PIC
  -> PIC ACK
  -> backend commit new state
  -> backend publish room_update
```

ถ้า `PIC ไม่ ACK`:

- ห้ามเปลี่ยน room state
- ห้ามปล่อย room_update แบบสำเร็จ
- ต้องสร้าง command failure หรือ alert event

---

## 6. Anti-Fraud Triggers

ระบบควรสร้าง alert เมื่อ:

- มี `open_room` โดยไม่มี `check_in`
- มี `extend_stay` แต่ไม่มี payment
- command เปิด relay สำเร็จในระบบ แต่ feedback จริงไม่ตรง
- room แสดง `vacant` แต่ sensor บอกว่ามีการใช้งาน
- handheld เปิดห้องจากอุปกรณ์หรือผู้ใช้ที่ไม่มีสิทธิ์

---

## 8. State Record Shape

room state ที่ backend ควรถืออย่างน้อย:

```json
{
  "room_id": "1205",
  "status": "occupied",
  "stay_mode": "overnight",
  "case": "overnight",
  "remaining_minutes": 95,
  "controller_online": true
}
```

---

## 7. Offline Handling

```text
normal state
  -> transport/controller timeout
controller_offline
  -> reconnect + snapshot sync
previous or recalculated state
```

กติกา:

- offline เป็น transport condition ไม่ใช่ business success
- หลัง reconnect backend ต้องขอ snapshot แล้วคำนวณ state ใหม่
