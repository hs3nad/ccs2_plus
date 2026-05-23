# CCS-2PLUS PRO Manual QA Script

เอกสารนี้ใช้สำหรับ manual QA ของ `flutter-dashboard`, `flutter-handheld`, และ `flutter-cashier`

เป้าหมาย:

- ยืนยันว่า `REST + SSE` ทำงานครบ flow
- ยืนยันว่า room state, audit, และ alert sync กันระหว่างแอป
- ยืนยันว่า pending / delayed / retry ใน frontend ทำงานตาม convention

---

## 1. Test Setup

ต้องเตรียม:

- `go-backend` รันอยู่และเข้าถึงได้
- `flutter-dashboard`
- `flutter-handheld`
- `flutter-cashier`

แนะนำให้ทุกแอปรันด้วย config เดียวกัน:

```bash
flutter run \
  --dart-define=BACKEND_BASE_URL=http://localhost:8080 \
  --dart-define=PENDING_TIMEOUT_SECONDS=6
```

หมายเหตุ:

- ถ้าเปลี่ยน timeout เป็นค่าอื่น ให้ใช้ค่านั้นแทนใน expected result ของ delayed cases

---

## 2. Global Smoke Test

### Case G1: Backend Reachability

steps:

1. เปิดทั้ง 3 แอป
2. ตรวจว่าไม่มี error banner ตอนเริ่ม

expected:

- ทุกแอปโหลดขึ้นได้
- room list/room board แสดงข้อมูลจาก backend

### Case G2: Initial Load + Full State

steps:

1. เปิดแอปใหม่จาก cold start
2. สังเกตรายการห้องในแต่ละแอป

expected:

- ทุกแอปมี room list จาก `GET /api/state`
- หลังต่อ `SSE` แล้ว state ไม่หาย

---

## 3. Dashboard QA

### Case D1: Room Board Realtime Update

steps:

1. เปิด `dashboard`
2. ใช้ `cashier` ทำ `check_in` ห้องหนึ่ง
3. กลับมาดู `dashboard`

expected:

- room tile เปลี่ยนเป็น `occupied`
- `stay_mode` และ `case` อัปเดตตาม
- ไม่ต้องกด refresh

### Case D2: Audit Event Realtime

steps:

1. เปิด `dashboard` ไปที่ section `Audit`
2. ใช้ `handheld` กด `open_room`

expected:

- มีรายการใหม่เข้า audit list ด้านบนสุด
- action ต้องตรงกับ `open_room`
- actor/result/detail แสดงได้ตาม payload

### Case D3: Fraud Alert Realtime

steps:

1. เปิด `dashboard` ไปที่ section `Anti-Fraud`
2. เลือกห้องที่ยัง `vacant`
3. ใช้ `handheld` กด `open_room`

expected:

- มี alert ใหม่เข้า list โดยไม่ต้อง refresh
- rule ต้องเป็น `OPEN_ROOM_WHILE_VACANT`

---

## 4. Handheld QA

### Case H1: Open Room Happy Path

steps:

1. เปิด `handheld`
2. เลือกห้อง `vacant`
3. กด `Open Room`

expected:

- ปุ่มเปลี่ยนเป็น `Sending...`
- ถ้า SSE กลับมาทันเวลา pending ต้องหาย
- dashboard audit ต้องมี `open_room`

### Case H2: Service Happy Path

steps:

1. เลือกห้อง `occupied`
2. กด `Service`
3. เลือก service item และ confirm

expected:

- UI ขึ้น `Sending...`
- เมื่อ SSE กลับมา pending ต้องหาย
- dashboard audit ต้องมี `service`

### Case H3: Start Cleaning Happy Path

steps:

1. เลือกห้องที่เหมาะกับ cleaning
2. กด `Start Cleaning`

expected:

- ห้องเปลี่ยนเป็น `cleaning`
- dashboard และ cashier เห็นสถานะใหม่โดยไม่ต้อง refresh

### Case H4: Delayed State

steps:

1. จำลองกรณี backend รับ REST แล้วไม่ส่ง `room_update` กลับในเวลาที่กำหนด
2. กด action เช่น `Open Room`
3. รอเกิน `PENDING_TIMEOUT_SECONDS`

expected:

- UI เปลี่ยนจาก `Sending...` เป็น `Delayed > Ns`
- ข้อความบอกว่ากด retry ได้
- room เดิมยังไม่ถูกล้างทิ้ง

### Case H5: Retry After Delayed

steps:

1. อยู่ในสถานะ delayed ตาม case H4
2. กดปุ่ม retry จากปุ่มเดิม

expected:

- action ถูกส่งใหม่ได้
- ถ้า SSE กลับมาหลัง retry pending/delayed ต้องหาย

### Case H6: Delayed Then Recover Without Retry

steps:

1. รัน mock scenario `delayed-room-update-recover`
2. กด action จาก `handheld` เช่น `Open Room`
3. รอให้ UI เปลี่ยนจาก `Sending...` เป็น `Delayed > Ns`
4. รอต่อจน mock ส่ง `room_update` กลับมา

expected:

- UI เข้าสถานะ `Delayed > Ns` ก่อน
- โดยไม่ต้องกด retry สถานะ delayed ต้องหายเองเมื่อ `room_update` กลับมา
- room card และ room detail ต้องอัปเดตตาม state ใหม่

---

## 5. Cashier QA

### Case C1: Check-In Happy Path

steps:

1. เปิด `cashier`
2. เลือกห้อง `vacant`
3. กด `Check-In`
4. เลือก `stay_mode` และ confirm

expected:

- ปุ่มขึ้น `Sending...`
- เมื่อ SSE กลับมา room tile/detail/billing panel อัปเดต
- dashboard เห็นห้องเป็น `occupied`

### Case C2: Check-Out Happy Path

steps:

1. เลือกห้อง `occupied`
2. กด `Check-Out`

expected:

- ปุ่มขึ้น `Sending...`
- เมื่อ SSE กลับมา room เปลี่ยนเป็น `vacant`

### Case C3: Extend Stay Happy Path

steps:

1. เลือกห้อง `occupied`
2. กด `Extend Stay`
3. กรอกข้อมูลแล้ว confirm

expected:

- panel แสดง pending ทันที
- เมื่อ SSE กลับมา `remaining_minutes` อัปเดต
- dashboard room board เปลี่ยนตาม

### Case C4: Receive Payment Happy Path

steps:

1. เลือกห้อง `occupied`
2. กด `Receive Payment`
3. กรอกข้อมูลและ confirm

expected:

- panel ขึ้น `Sending...`
- เมื่อ SSE กลับมา pending หาย
- dashboard audit มี `receive_payment`

### Case C5: Delayed + Retry

steps:

1. จำลองกรณีไม่มี `room_update` กลับภายใน timeout
2. ยิง `check_in`, `check_out`, `extend`, หรือ `payment`
3. รอเกิน timeout
4. กด retry

expected:

- UI ขึ้น `Delayed > Ns`
- ปุ่มเดิมกด retry ได้
- เมื่อ SSE กลับมาหลัง retry สถานะต้อง clear

### Case C6: Delayed Then Recover Without Retry

steps:

1. รัน mock scenario `delayed-room-update-recover`
2. ยิง `check_in`, `check_out`, `extend`, หรือ `payment`
3. รอเกิน timeout ให้ UI ขึ้น `Delayed > Ns`
4. ไม่ต้องกด retry และรอ `room_update` กลับมา

expected:

- room tile, detail panel, และ billing panel แสดง delayed ก่อน
- เมื่อ `room_update` กลับมา delayed ต้องหายเอง
- state ใหม่ต้องตรงกับ action ที่ส่งไป

---

## 6. Cross-App Sync QA

### Case X1: Handheld -> Dashboard

steps:

1. ใช้ `handheld` ยิง `open_room`
2. ดู `dashboard`

expected:

- audit list อัปเดต
- fraud alert อัปเดตตาม rule ที่เข้าเงื่อนไข

### Case X2: Cashier -> Dashboard

steps:

1. ใช้ `cashier` ยิง `check_in` หรือ `extend`
2. ดู `dashboard`

expected:

- room board อัปเดต
- audit list อัปเดต

### Case X3: Handheld -> Cashier

steps:

1. ใช้ `handheld` ยิง `start_cleaning`
2. ดู `cashier`

expected:

- room status บน `cashier` เปลี่ยนตาม realtime

---

## 7. Failure QA

### Case F1: REST Failure

steps:

1. ทำให้ backend ตอบ error หรือปิด backend ชั่วคราว
2. กด action จาก `handheld` หรือ `cashier`

expected:

- UI แสดง error banner หรือข้อความผิดพลาด
- pending ไม่ค้าง
- state เก่าไม่หาย

### Case F2: SSE Disconnect

steps:

1. ทำให้ stream `/events` หลุด
2. สังเกต frontend

expected:

- UI แสดง error ตาม store
- room state ล่าสุดยังอยู่
- ถ้ากลับมาต่อใหม่ได้ state ต้อง sync ต่อได้

### Case F3: Delayed Recover Scenario

steps:

1. รัน mock ด้วย `DELAYED_RECOVER_AFTER=8s` หรือ `10s`
2. ใช้ `handheld` หรือ `cashier` ยิง action ใด action หนึ่ง
3. จับเวลาตั้งแต่กด action จน delayed ขึ้น
4. จับเวลาจน `room_update` กลับมา

expected:

- delayed ต้องขึ้นหลัง `PENDING_TIMEOUT_SECONDS`
- `room_update` ต้องกลับมาภายหลังตามค่าที่ mock ตั้งไว้
- UI ต้อง recover เองโดยไม่ต้อง reload หน้า

---

## 8. Exit Criteria

ถือว่าผ่านรอบ manual QA นี้เมื่อ:

- happy path หลักของทั้ง 3 แอปผ่าน
- room state sync กันระหว่างแอป
- `audit_event` และ `fraud_alert` realtime ทำงานใน dashboard
- `Sending...`, `Delayed > Ns`, และ retry ทำงานใน handheld/cashier
- delayed-recover flow ทำงานโดยไม่ต้องกด retry
- REST error และ SSE disconnect ไม่ทำให้ UI state พัง

---

## 9. References

- [docs/current_architecture.md](/home/kaina/work/project/ccs2_plus/docs/current_architecture.md)
- [docs/api_contract.md](/home/kaina/work/project/ccs2_plus/docs/api_contract.md)
- [docs/frontend_conventions.md](/home/kaina/work/project/ccs2_plus/docs/frontend_conventions.md)
