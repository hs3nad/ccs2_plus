# CCS-2PLUS PRO Frontend Conventions

เอกสารนี้สรุป convention กลางของ `flutter-dashboard`, `flutter-handheld`, และ `flutter-cashier`

---

## 1. Transport Rule

- ใช้ `REST` สำหรับ initial load และ user command
- ใช้ `SSE` สำหรับ realtime update หลัง state เปลี่ยน
- frontend ต้องไม่คุยกับ hardware โดยตรง
- backend คือ source of truth เสมอ

flow หลัก:

```text
UI action
-> REST command to go-backend
-> backend accepts
-> SSE event returns
-> UI updates from SSE
```

---

## 2. State Loading Rule

- ตอนเริ่มแอป ให้เรียก `GET /api/state`
- หลังจากนั้นให้ต่อ `GET /events`
- เมื่อได้ `full_state` ให้ replace room list ทั้งก้อน
- เมื่อได้ `room_update` ให้ patch เฉพาะห้องที่เปลี่ยน

สำหรับ `dashboard`:

- `audit_event` ใช้ append audit list แบบ realtime
- `fraud_alert` ใช้ append alert list แบบ realtime

---

## 3. Command Rule

action naming ใน frontend ต้องยึดตาม backend contract:

- `check_in`
- `check_out`
- `extend_stay`
- `receive_payment`
- `open_room`
- `service`
- `start_cleaning`
- `finish_cleaning`

กติกา:

- ใช้ `snake_case` ตรงกับ REST/SSE/backend
- หลีกเลี่ยง alias ฝั่ง UI ที่ทำให้ชื่อ action ไม่ตรงกับ backend

---

## 4. Pending And Delayed Rule

สำหรับ `flutter-handheld` และ `flutter-cashier`:

- เมื่อกด action ให้ขึ้น `Sending...` ทันที
- ไม่ต้อง `refresh()` หลังทุก action
- รอ `room_update` จาก SSE เป็นตัวเคลียร์ pending
- ถ้าเกิน timeout ที่กำหนดและยังไม่มี SSE ให้เปลี่ยนเป็น `Delayed > Ns`
- เมื่ออยู่ในสถานะ delayed ให้ผู้ใช้กด retry จากปุ่มเดิมได้

กติกา UX:

- `Sending...` = command ถูกส่งแล้วและกำลังรอ backend/SSE
- `Delayed > Ns` = command ยังไม่ถูก confirm ในเวลาที่คาดไว้
- `error banner` = request fail ที่ REST level หรือ SSE connection fail

---

## 5. Config Rule

ทุกแอปต้องรองรับ `dart-define` ขั้นต่ำดังนี้:

- `BACKEND_BASE_URL`
- `PENDING_TIMEOUT_SECONDS`

ตัวอย่าง:

```bash
flutter run \
  --dart-define=BACKEND_BASE_URL=http://localhost:8080 \
  --dart-define=PENDING_TIMEOUT_SECONDS=6
```

หมายเหตุ:

- `dashboard` ยังไม่ได้ใช้ pending/delayed UI แบบเต็ม แต่คง `PENDING_TIMEOUT_SECONDS` ไว้เพื่อให้ convention ตรงกันทั้ง 3 แอป
- `flutter analyze` ไม่รองรับ `--dart-define`
- ใช้ `flutter analyze` แบบปกติ และใช้ `--dart-define` เฉพาะตอน `run` หรือ `build`

---

## 6. Error Handling Rule

- error จาก REST ให้แสดงผ่าน banner หรือข้อความกลางจอ
- error จาก SSE connection ให้เก็บไว้ใน store และ notify UI
- อย่าล้าง room state เดิมทิ้งเพียงเพราะ SSE หลุดชั่วคราว

---

## 7. Scope Today

สิ่งที่ใช้จริงแล้ว:

- `REST + SSE`
- `full_state`
- `room_update`
- `audit_event`
- `fraud_alert`
- pending / delayed state ใน `handheld` และ `cashier`

สิ่งที่ยังไม่ปิด:

- `flutter analyze` ยังไม่ได้รันใน sandbox นี้
- Flutter widget tests ยังไม่ได้เพิ่ม

---

## 8. References

- [docs/current_architecture.md](/home/kaina/work/project/ccs2_plus/docs/current_architecture.md)
- [docs/api_contract.md](/home/kaina/work/project/ccs2_plus/docs/api_contract.md)
- [flutter-dashboard/README.md](/home/kaina/work/project/ccs2_plus/flutter-dashboard/README.md)
- [flutter-handheld/README.md](/home/kaina/work/project/ccs2_plus/flutter-handheld/README.md)
- [flutter-cashier/README.md](/home/kaina/work/project/ccs2_plus/flutter-cashier/README.md)

---

## 9. Pre-Merge Checklist

- รัน `flutter analyze` ในแอปที่แก้
- เช็กว่า `BACKEND_BASE_URL` และ `PENDING_TIMEOUT_SECONDS` ถูกตั้งค่าตรงกับ environment ที่ใช้ทดสอบ
- เปิดแอปแล้ว confirm ว่า initial load จาก `GET /api/state` ทำงานปกติ
- ต่อ `SSE` แล้ว confirm ว่า `full_state` เข้าและ room list ถูก replace ได้
- ยิง action อย่างน้อย 1 ครั้งแล้ว confirm ว่า `room_update` กลับมาและ UI เปลี่ยนโดยไม่ต้อง manual refresh
- สำหรับ `dashboard` ให้เช็กว่า `audit_event` และ `fraud_alert` เด้งเข้า list ได้จริง
- สำหรับ `handheld` และ `cashier` ให้เช็กว่า action ปกติขึ้น `Sending...`
- สำหรับ `handheld` และ `cashier` ให้เช็ก delayed flow โดยจำลองกรณีไม่มี `room_update` เกิน timeout แล้ว UI เปลี่ยนเป็น `Delayed > Ns`
- หลัง delayed ให้เช็กว่าปุ่ม retry กดได้และ state ไม่ค้าง
- เช็กว่า error banner แยกกรณี REST fail ออกจาก delayed/retry ได้ชัด
