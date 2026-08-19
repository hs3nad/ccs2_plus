# CCS-2PLUS PRO Mock Backend Scenarios

เอกสารนี้อธิบาย mock backend สำหรับใช้ทดสอบ frontend ในเคสที่ทำกับ backend ปกติได้ยาก เช่น `delayed` และ `SSE drop`

---

## 1. Purpose

mock server นี้ใช้สำหรับ:

- จำลอง `room_update` ที่หายไปหรือมาช้า
- จำลอง `SSE` ที่หลุดหลังส่ง `full_state`
- ให้ tester เดินเคส `Sending...`, `Delayed > Ns`, และ retry ได้ง่าย

ตัว mock นี้แยกจาก `go-backend` หลัก และไม่กระทบ flow production

---

## 2. Run

ใช้ script นี้:

```bash
./scripts/mock_backend_scenario.sh normal
```

หรือ:

```bash
./scripts/mock_backend_scenario.sh delayed-room-update
./scripts/mock_backend_scenario.sh delayed-room-update-recover
./scripts/mock_backend_scenario.sh sse-drop-after-full-state
```

ค่า default:

- listen ที่ `:18080`

ถ้าต้องการเปลี่ยน port:

```bash
LISTEN_ADDR=:18081 ./scripts/mock_backend_scenario.sh delayed-room-update
```

ถ้าต้องการทดสอบ delayed แล้ว recover หลัง `8-10s`:

```bash
DELAYED_RECOVER_AFTER=8s ./scripts/mock_backend_scenario.sh delayed-room-update-recover
DELAYED_RECOVER_AFTER=10s ./scripts/mock_backend_scenario.sh delayed-room-update-recover
```

จากนั้นให้ Flutter app ชี้ไปที่ mock backend:

```bash
flutter run \
  --dart-define=BACKEND_BASE_URL=http://localhost:18080 \
  --dart-define=PENDING_TIMEOUT_SECONDS=6
```

---

## 3. Supported Scenarios

### `normal`

behavior:

- ตอบ `REST` ปกติ
- ส่ง `full_state`
- ส่ง `room_update`
- ส่ง `audit_event`
- ส่ง `fraud_alert` เมื่อเข้าเงื่อนไข

เหมาะกับ:

- smoke test frontend
- cross-app sync test

### `delayed-room-update`

behavior:

- ตอบ `REST` เป็น `accepted`
- ส่ง `audit_event` ได้
- ส่ง `fraud_alert` ได้
- ไม่ส่ง `room_update`

เหมาะกับ:

- ทดสอบ `Sending... -> Delayed > Ns`
- ทดสอบ retry flow ใน `handheld` และ `cashier`

### `delayed-room-update-recover`

behavior:

- ตอบ `REST` เป็น `accepted`
- ส่ง `audit_event` ได้ทันที
- ส่ง `fraud_alert` ได้ทันทีเมื่อเข้าเงื่อนไข
- หน่วง `room_update` ก่อนส่งกลับ
- ค่า default ของ delay คือ `8s`
- ปรับเป็น `10s` หรือค่าอื่นได้ผ่าน `DELAYED_RECOVER_AFTER`

เหมาะกับ:

- ทดสอบ `Sending... -> Delayed > Ns -> recover`
- ทดสอบว่า UI clear delayed เองเมื่อ `room_update` กลับมา
- ทดสอบ flow ที่ไม่ต้องกด retry

### `sse-drop-after-full-state`

behavior:

- client ต่อ `GET /events` ได้
- server ส่ง `full_state` หนึ่งครั้ง
- จากนั้นปิด stream

เหมาะกับ:

- ทดสอบการรับมือ `SSE disconnect`
- เช็กว่า UI ไม่ล้าง state เดิมทิ้งเมื่อ stream หลุด

---

## 4. Endpoints

mock server รองรับ endpoint หลักที่ frontend ใช้ตอนนี้:

- `GET /ping`
- `GET /api/state`
- `GET /api/audit/events`
- `GET /api/alerts`
- `POST /api/rooms/{room_id}/checkin`
- `POST /api/rooms/{room_id}/checkout`
- `POST /api/rooms/{room_id}/extend`
- `POST /api/rooms/{room_id}/payment`
- `POST /api/rooms/{room_id}/command`
- `GET /events`

---

## 5. Tester Notes

- ใช้ `normal` เมื่อต้องการเช็ก happy path
- ใช้ `delayed-room-update` เมื่อต้องการเช็ก delayed/retry
- ใช้ `delayed-room-update-recover` เมื่อต้องการเช็ก delayed แล้วกลับมาหายเอง
- ใช้ `sse-drop-after-full-state` เมื่อต้องการเช็กกรณี stream หลุด
- ถ้าจะเดิน test script เต็ม ให้ใช้คู่กับ [docs/manual_qa_frontend.md](/home/kaina/work/project/ccs2_plus/docs/manual_qa_frontend.md)

---

## 6. References

- [docs/manual_qa_frontend.md](/home/kaina/work/project/ccs2_plus/docs/manual_qa_frontend.md)
- [docs/frontend_conventions.md](/home/kaina/work/project/ccs2_plus/docs/frontend_conventions.md)
