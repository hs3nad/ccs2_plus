# CCS-2PLUS PRO DB Model

เอกสารนี้สรุป model ระดับฐานข้อมูลที่รองรับ `stay_mode`, `case`, room mapping และ report แบบ legacy CCS2

---

## 1. Design Principles

- realtime state อาจอยู่ใน memory cache ได้
- แต่ข้อมูลธุรกิจและ audit ต้อง persist
- field ที่มาจากคู่มือเดิม เช่น `guest_in`, `guest_out`, `clean_up`, `used`, `case`, `remark` ต้องมีที่เก็บชัดเจน

---

## 2. Core Tables

### `rooms`

```sql
CREATE TABLE rooms (
    room_id              TEXT PRIMARY KEY,
    floor                INTEGER NOT NULL,
    controller_id        TEXT NOT NULL,
    rs485_address        INTEGER NOT NULL,
    card_index           INTEGER NOT NULL,
    port_index           INTEGER NOT NULL,
    display_enabled      INTEGER NOT NULL DEFAULT 1,
    remark               TEXT
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

---

## 3. Recommended Values

`stay_mode`:

- `overnight`
- `temporary`
- `monthly`
- `yearly`

`case_code`:

- เริ่มต้นให้ใช้ค่าเดียวกับ `stay_mode`
- ถ้าต้องรองรับ report เดิมมากขึ้น ค่อยแยก formatting layer ทีหลัง

---

## 4. Legacy Report Compatibility

field ที่ต้องดึงออกไปทำ report ได้:

- room
- guest_in_at
- guest_out_at
- clean_up_at
- used_minutes
- case_code
- remark
