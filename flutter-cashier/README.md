# Flutter Cashier

แอปนี้เป็น client สำหรับ cashier / front office workflow

scope ปัจจุบัน:

- คุยกับ `go-backend` ผ่าน `REST + SSE`
- เหมาะกับ front office flow, check-in/check-out, room status board
- ยังไม่คุยกับ hardware โดยตรง

architecture อ้างอิง:

- [current_architecture.md](/home/kaina/work/project/ccs2_plus/docs/current_architecture.md)
- [api_contract.md](/home/kaina/work/project/ccs2_plus/docs/api_contract.md)

run config:

- `BACKEND_BASE_URL` ใช้กำหนด URL ของ `go-backend`
- `PENDING_TIMEOUT_SECONDS` ใช้กำหนดเวลารอก่อน UI เปลี่ยนจาก `Sending...` เป็น `Delayed > Ns`

ตัวอย่าง:

```bash
flutter run \
  --dart-define=BACKEND_BASE_URL=http://localhost:8080 \
  --dart-define=PENDING_TIMEOUT_SECONDS=6
```

analyze:

```bash
flutter analyze
```

หมายเหตุ:

- `flutter analyze` ไม่รองรับ `--dart-define`
- ใช้ `--dart-define` เฉพาะตอน `flutter run` หรือ `flutter build`
