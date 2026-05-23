# Pi1 Board

| Item | Value |
|---|---:|
| Dashboard port | 6010 |
| SSH port | 6110 |

Public dashboard endpoint:

```text
http://34.87.151.202:6010/
```

Runtime transport:

- REST API: `/api/*`
- SSE stream: `/events`
- `frpc` remote dashboard/backend: `6010`
- `frpc` remote SSH: `6110`

## Deployment Goal

`raspi1` ในโปรเจกต์นี้ทำหน้าที่เป็น local server หน้างาน โดยรัน:

- `hotel-maid-backend` ที่ `0.0.0.0:6010`
- Flutter dashboard static files ที่ `/opt/hotel-dashboard`
- `frpc` เพื่อเปิด public endpoint `34.87.151.202:6010`
- `OpenSSH` เพื่อ remote ผ่าน `34.87.151.202:6110`

## Recommended Layout

แนะนำให้วางไฟล์บนเครื่องตามนี้:

```text
/usr/local/bin/hotel-maid-backend
/opt/hotel-dashboard/
├── index.html
├── assets/...
├── data/ccs2plus.db
└── exports/
/opt/frp/
├── frpc
└── ccs2plus-6010.toml
```

สิ่งที่ควรมีล่วงหน้า:

- Raspberry Pi OS ที่เปิด SSH แล้ว
- user สำหรับ runtime เช่น `kaina`
- network ออก internet ได้
- `frpc` binary สำหรับ ARM ของ Pi1
- service `ssh` ทำงานปกติบน port `22`

## Initial Setup

1. update package พื้นฐาน

```bash
sudo apt update
sudo apt install -y rsync curl
```

2. สร้างโฟลเดอร์ runtime

```bash
sudo mkdir -p /opt/hotel-dashboard/data /opt/hotel-dashboard/exports /opt/frp
sudo chown -R kaina:kaina /opt/hotel-dashboard /opt/frp
```

3. วาง binary และ dashboard

```bash
sudo cp /path/to/go-backend /usr/local/bin/hotel-maid-backend
sudo chmod +x /usr/local/bin/hotel-maid-backend
rsync -av --delete /path/to/dashboard-web/ /opt/hotel-dashboard/
```

4. วาง `frpc` และ config

```bash
cp /path/to/frpc /opt/frp/frpc
chmod +x /opt/frp/frpc
cp /path/to/ccs2plus-6010.toml /opt/frp/ccs2plus-6010.toml
```

5. ถ้ายังไม่มีฐานข้อมูลเริ่มต้น ให้คัดลอกจาก build/deploy package ไปไว้ที่
`/opt/hotel-dashboard/data/ccs2plus.db`

## Serial Port

ใช้ hardware UART ภายในของ Pi 1:

| Device | หมายเหตุ |
|---|---|
| `/dev/ttyAMA0` | hardware UART (PL011) — ใช้งานจริง |
| `/dev/serial0` | symlink → `/dev/ttyAMA0` |

> **ต้องตั้งค่าก่อนใช้:**
> 1. `sudo raspi-config` → Interface Options → Serial Port
>    - "login shell over serial" → **No**
>    - "serial port hardware enabled" → **Yes**
> 2. ถ้าใช้ Pi 1 ดั้งเดิม (ไม่มี Bluetooth) — `/dev/ttyAMA0` ว่างเสมอ ไม่ต้องทำอะไรเพิ่ม

ตรวจสอบหลังตั้งค่า:

```bash
ls -l /dev/ttyAMA0 /dev/serial0
```

Run command:

```bash
./go-backend \
  -listen 0.0.0.0:6010 \
  -serial /dev/ttyAMA0 \
  -baud 9600 \
  -db-path ./data/ccs2plus.db \
  -export-dir ./exports \
  -export-timezone Asia/Bangkok \
  -dashboard-dir ./dashboard-web
```

frpc config:

```bash
./bin/frpc verify -c ./frpc/ccs2plus-6010.toml
./bin/frpc -c ./frpc/ccs2plus-6010.toml
```

บนเครื่องจริง path ที่ใช้งานคือ:

```bash
/opt/frp/frpc verify -c /opt/frp/ccs2plus-6010.toml
sudo /opt/frp/frpc -c /opt/frp/ccs2plus-6010.toml
```

## systemd Services

template files ใน repo:

- [hotel-maid-backend.service](/home/kaina/work/project/ccs2_plus/boards/raspi1/systemd/hotel-maid-backend.service:1)
- [frpc.service](/home/kaina/work/project/ccs2_plus/boards/raspi1/systemd/frpc.service:1)
- [journald.conf](/home/kaina/work/project/ccs2_plus/boards/raspi1/systemd/journald.conf:1)

backend service:

```ini
[Unit]
Description=CCS-2PLUS Backend
After=network.target

[Service]
ExecStart=/usr/local/bin/hotel-maid-backend \
  -listen 0.0.0.0:6010 \
  -serial /dev/ttyAMA0 \
  -baud 9600 \
  -db-path /opt/hotel-dashboard/data/ccs2plus.db \
  -export-dir /opt/hotel-dashboard/exports \
  -export-timezone Asia/Bangkok \
  -dashboard-dir /opt/hotel-dashboard
Restart=always
RestartSec=5
User=kaina

[Install]
WantedBy=multi-user.target
```

frpc service:

```ini
[Unit]
Description=FRP Client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/opt/frp/frpc -c /opt/frp/ccs2plus-6010.toml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

ติดตั้งจาก template:

```bash
sudo cp boards/raspi1/systemd/hotel-maid-backend.service /etc/systemd/system/
sudo cp boards/raspi1/systemd/frpc.service /etc/systemd/system/
```

Enable:

```bash
sudo systemctl daemon-reload
sudo systemctl enable hotel-maid-backend frpc
sudo systemctl restart hotel-maid-backend frpc
```

Logs:

```bash
journalctl -u hotel-maid-backend -f
journalctl -u frpc -f
```

ถ้าไม่ต้องการเขียน log ลง SD card ให้ตั้ง `journald` เป็น `Storage=volatile`
เพื่อเก็บ log ไว้ใน `/run/log/journal` ซึ่งอยู่บน `tmpfs`

ตัวอย่าง `/etc/systemd/journald.conf`:

```ini
[Journal]
Storage=volatile
RuntimeMaxUse=32M
SystemMaxUse=0
```

ติดตั้งจาก template:

```bash
sudo cp boards/raspi1/systemd/journald.conf /etc/systemd/journald.conf
```

จากนั้นใช้:

```bash
sudo systemctl restart systemd-journald
sudo rm -rf /var/log/journal
```

## Build And Deploy From Repo

บนเครื่อง dev:

```bash
./scripts/build_raspi1.sh all
PI_HOST=<raspi-ip> ./scripts/deploy_raspi1.sh all
```

สิ่งที่สคริปต์ deploy ทำ:

- ส่ง `dist/raspi1/go-backend` ไป `/usr/local/bin/hotel-maid-backend`
- ส่ง `dist/raspi1/dashboard-web/` ไป `/opt/hotel-dashboard`
- restart service `hotel-maid-backend`

หมายเหตุ:

- สคริปต์ deploy ปัจจุบันยังไม่ deploy `frpc` binary หรือ `frpc.service`
- ควรติดตั้ง `/opt/frp/frpc`, config และ `frpc.service` ครั้งแรกด้วยตนเองก่อน

ถ้าจะติดตั้ง template ฝั่ง `systemd` และ `journald` แบบครั้งเดียวจาก repo ใช้:

```bash
sudo ./scripts/setup_raspi1.sh
```

ถ้าต้องการติดตั้งไฟล์ก่อนโดยยังไม่ restart service ใช้:

```bash
sudo ./scripts/setup_raspi1.sh --no-restart
```

## Verification Checklist

เช็กในเครื่อง:

```bash
sudo systemctl status hotel-maid-backend
sudo systemctl status frpc
curl http://127.0.0.1:6010/api/state
```

เช็กจากภายนอก:

```bash
curl http://34.87.151.202:6010/api/state
ssh -p 6110 kaina@34.87.151.202
```

เช็ก logging mode:

```bash
ls -ld /run/log/journal
mount | grep ' /run '
grep -E '^[# ]*Storage=' /etc/systemd/journald.conf
```

ผลที่ต้องการ:

- `hotel-maid-backend` เป็น `active (running)`
- `frpc` เป็น `active (running)`
- local `curl` ตอบกลับได้
- public `curl` ตอบกลับได้
- `/run` เป็น `tmpfs`
- `Storage=volatile`

## Reboot Test

หลังตั้งค่าครบ ให้ทดสอบ 1 รอบ:

```bash
sudo reboot
```

หลังบูตกลับมา:

```bash
sudo systemctl status hotel-maid-backend
sudo systemctl status frpc
curl http://127.0.0.1:6010/api/state
curl http://34.87.151.202:6010/api/state
```

ถ้าทุกคำสั่งผ่าน ถือว่า `raspi1` พร้อมใช้งานกับโปรเจกต์นี้

## Build

```bash
./scripts/build_raspi1.sh all
```

## Deploy

```bash
PI_HOST=<ip> ./scripts/deploy_raspi1.sh all
```

## Artifacts

```text
dist/raspi1/go-backend
dist/raspi1/dashboard-web/
dist/raspi1/build-info.txt
```
