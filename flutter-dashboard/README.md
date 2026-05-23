# Flutter Dashboard

แอปนี้เป็น dashboard สำหรับ front office และ manager

scope ปัจจุบัน:

- คุยกับ `go-backend` ผ่าน `REST + SSE`
- ใช้ดู room board, status, stay mode, case
- ใช้ส่ง action หลัก เช่น `check_in`, `check_out`, `start_cleaning`, `finish_cleaning`

architecture อ้างอิง:

- [current_architecture.md](/home/kaina/work/project/ccs2_plus/docs/current_architecture.md)
- [api_contract.md](/home/kaina/work/project/ccs2_plus/docs/api_contract.md)

run config:

- `BACKEND_BASE_URL` ใช้กำหนด URL ของ `go-backend`
- `PENDING_TIMEOUT_SECONDS` เก็บ convention เดียวกับแอปอื่น แม้ dashboard ปัจจุบันยังไม่ได้ใช้ pending/delayed UI แบบ `handheld` และ `cashier`

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
