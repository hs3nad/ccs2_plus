# CCS-2PLUS PRO Legacy Protocol Reference

เอกสารนี้สรุป `legacy protocol` จาก master code เดิม และแยกบทบาทของอุปกรณ์ให้ชัดเจนตาม source code ที่มีอยู่จริง

บริบทที่ใช้สรุป:

- `master` คือเครื่องควบคุมหลักเดิม
- `PIC slave` คือบอร์ดควบคุม output 16 ช่อง
- `PIC display` คือบอร์ด LED indicator / blink display

ในงานใหม่:

- PIC ยังมีอยู่ แต่แยกอยู่ในส่วน `slave` และ `display`
- backend จะรับ business logic แทน master เดิม
- ปกติ backend หรือเครื่อง local controller จะต่อ `RS-485` ไป PIC ได้โดยตรง
- `esp32-gateway` เป็น `optional transport adapter` ใช้เมื่อระยะสาย RS-485 ยาวเกินไป หรืออยากแยกเป็น gateway ตามโซน/ชั้น

---

## 0. Default Deployment

deployment หลักของระบบตอนนี้คือ:

```text
Handheld / Tablet
        |
        | REST + SSE
        v
Raspberry Pi 1
  |- Go Backend
  |- Flutter Dashboard
  |- frpc
  |- RS-485 service / USB-RS485 adapter
        |
        v
PIC Slave Board(s) + PIC Display Board(s)
```

หลักการ:

- `Raspberry Pi 1` เป็น local server หลัก
- `backend` และ `dashboard` รันบน Pi1
- `handheld/tablet` คุยกับ backend เท่านั้น
- `Pi1` ส่ง `legacy frame` ไป PIC ผ่าน `RS-485` ตรง
- `MQTT` ไม่ใช้ในโหมดนี้
- `ESP32 Gateway` ใช้เฉพาะกรณี optional deployment

### 0.1 Runtime Responsibilities In Default Deployment

| Component | Responsibility |
|---|---|
| `Go Backend` | auth, room state, audit, anti-fraud, action orchestration |
| `Flutter Dashboard` | front office / manager UI |
| `Handheld/Tablet` | room operation UI |
| `RS-485 transport service` | encode legacy frame, send, retry, timeout |
| `PIC Slave` | latch relay/output |
| `PIC Display` | render LED / blink status |

### 0.2 Command Flow In Default Deployment

```text
Handheld / Dashboard
  -> REST command to backend
  -> backend validates and resolves room mapping
  -> backend builds desired room state
  -> RS-485 transport encodes legacy frames
  -> send slave frame
  -> send display frame
  -> backend records result and publishes room_update
```

### 0.3 Optional Deployment

ถ้าสาย `RS-485` ยาวเกินไปหรือแบ่งโซนหน้างานยาก ค่อยเปลี่ยน transport layer เป็น:

```text
Pi1 -> MQTT -> ESP32 Gateway -> RS-485 -> PIC
```

### 0.4 Legacy Master Capabilities From Manual

อ้างอิงจาก [คู่มือ ccs2.pdf](/home/kaina/work/project/ccs2_plus/คู่มือ%20ccs2.pdf)

คู่มือ master เดิมยืนยันความสามารถหลักของระบบดังนี้:

| Capability | Evidence from manual | Meaning for new system |
|---|---|---|
| Front-office control | มีคำอธิบาย `CCS FRONT OFFICE CHECK-IN CHECK-OUT SYSTEM` | dashboard/backend ใหม่ต้องรองรับ workflow หน้าเคาน์เตอร์ |
| Local time setup | มีเมนู `Set date/time` และ password | เวลาเป็นแกนหลักของ event log และ report |
| Room status board | มีหน้าแสดงสถานะห้องพักและแผงไฟโชว์ | สถานะห้องเป็นศูนย์กลางของ operator workflow |
| Room power control | มี flow `เช็คอิน`, `เช็คเอาท์`, `ชั่วคราว`, `ต่อเวลา` | action ใหม่ต้องคง semantic เดิมเหล่านี้ |
| Reporting/printing | มีเมนู `PERIOD`, `RECORD`, `SUMMARY`, `INFO` | backend ใหม่ควรเก็บข้อมูลพอสำหรับ report 4 แบบ |
| Historical persistence | คู่มือระบุว่าเก็บข้อมูลแม้ไฟดับ | backend ใหม่ควร persist event/audit data |
| Large room count | คู่มือระบุว่า master 1 ชุดควบคุมได้ไม่จำกัดจำนวนห้องพัก | architecture ใหม่ควร scale room mapping ได้ |
| Display integration | มีแผงไฟโชว์สถานะห้องและตู้รีเลย์เป็นองค์ประกอบหลัก | ต้องรักษา relay/display split ใน transport layer |
| Installation topology | มี wiring diagram ระหว่าง master, relay cabinet, display board และ room controller | โครงสร้างใหม่ควรแยก deployment logic ออกจาก business logic |

จากคู่มือเดิม สถานะ/โหมดใช้งานที่ยืนยันได้ชัดคือ:

- `เช็คอิน`
- `เช็คเอาท์`
- `ชั่วคราว`
- `ต่อเวลา`
- `ค้างคืน`

ความหมายต่อระบบใหม่:

- `check_in`, `check_out`, `extend_stay` เป็น first-class actions
- `temporary stay` และ `overnight stay` ควรเก็บเป็น booking/rental mode ใน backend
- report model ต้องรองรับ `guest_in`, `guest_out`, `clean_up`, `used`, `case`, `remark`

ข้อมูลที่คู่มือชี้ว่าควรเก็บใน backend ใหม่:

- room number
- guest in time
- guest out time
- clean up time
- used duration
- case / stay mode
- remark

ดูสรุปหน้า-by-หน้าที่ [docs/legacy_master_manual.md](/home/kaina/work/project/ccs2_plus/docs/legacy_master_manual.md)

---

## 1. Legacy System Topology

```text
Legacy Master
  |- keypad / LCD / printer / RTC
  |- room mapping
  |- guest state / overtime / cleanup logic
  |- RS485 master
  |
  +--> PIC Slave Board(s)
  |      |- 16 outputs per card
  |      |- DIP-switch card address
  |      |- write-only style command consumption
  |
  +--> PIC Display Board(s)
         |- LED status display
         |- blink patterns
         |- card/port to LED matrix mapping
```

---

## 2. Device Roles In Legacy Protocol

| Device | Role | Direction |
|---|---|---|
| Master | Generate room state and output commands | send |
| PIC Slave | Apply 16-channel output state | receive |
| PIC Display | Show LED status / blink pattern | receive |

หมายเหตุ:

- จาก code ที่อ่าน ฝั่ง slave และ display ไม่ได้ถือ business state โรงแรม
- state หลัก เช่น `IN`, `OUT`, `CLEAN`, `OT` ถูกเก็บและคำนวณที่ master
- PIC รับคำสั่งแล้วแปลงเป็น output หรือ LED state เท่านั้น

---

## 3. Master To PIC Slave Protocol (AMP_V2D Generation)

อ้างอิงหลักจาก:

- [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:2193)
- [SLAVE_V2_0/MAIN.C](/home/kaina/work/project/CCS/SLAVE_V2_0/MAIN.C:234)

### 3.1 Frame Shape

```text
':' ADDR DATA_LO DATA_HI CHKSUM
```

จำนวนรวม:

- 1 start byte
- 1 address byte
- 2 data bytes
- 1 checksum byte

รวมทั้งหมด 5 bytes ที่ master ส่งจริงจาก `Out485_16()`

### 3.2 Field-By-Field

| Offset | Field | Size | Description |
|---|---|---|---|
| 0 | `':'` | 1 byte | start-of-frame, ค่า `0x3A` |
| 1 | `ADDR` | 1 byte | slave/card address |
| 2 | `DATA_LO` | 1 byte | low byte ของ relay bitmask 16 บิต |
| 3 | `DATA_HI` | 1 byte | high byte ของ relay bitmask 16 บิต |
| 4 | `CHKSUM` | 1 byte | two's complement checksum |

### 3.3 Checksum

checksum คำนวณแบบ:

```text
sum = ':' + ADDR + DATA_LO + DATA_HI
checksum = (~sum) + 1
```

อ้างอิงจาก [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:2200) และ [SLAVE_V2_0/MAIN.C](/home/kaina/work/project/CCS/SLAVE_V2_0/MAIN.C:248)

หมายเหตุ:

- ใน code ฝั่ง slave มีร่องรอยว่าตั้งใจเช็ก checksum
- แต่เงื่อนไข check checksum ถูกคอมเมนต์ไว้ และเหลือเช็กแค่ address ตรงกับ `CardID`
- ดู [SLAVE_V2_0/MAIN.C](/home/kaina/work/project/CCS/SLAVE_V2_0/MAIN.C:261)

### 3.4 Addressing

ฝั่ง slave อ่าน address จาก DIP switch 4 บิต:

```text
CardID = (~P1) & 0x0f
```

ดังนั้น card address รองรับประมาณ `0..15`

อ้างอิงจาก [SLAVE_V2_0/MAIN.C](/home/kaina/work/project/CCS/SLAVE_V2_0/MAIN.C:231)

### 3.5 Output Encoding

ใน `AMP_V2D.C` master ไม่ได้ส่งสถานะห้อง 16 ช่องแบบ nibble-by-nibble แต่ส่ง `unsigned int dat` ที่เป็น relay bitmask 16 บิตไปตรง ๆ

ตัวอย่างจาก master:

- เริ่มที่ `out = 0x0001`
- `out <<= port`
- ถ้าสถานะห้องเป็น `2..4` ให้ `ROOM_OUTPUT[card] |= out`
- ถ้าสถานะห้องเป็น `1` ให้ `ROOM_OUTPUT[card] &= ~out`
- แล้วส่ง `Out485_16(card, ROOM_OUTPUT[card])`

อ้างอิงจาก [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:1656)

ดังนั้นในรุ่นนี้:

- `port 0` = bit 0
- `port 1` = bit 1
- ...
- `port 15` = bit 15

และ relay effect คือ:

- bit = `1` -> relay on
- bit = `0` -> relay off

### 3.6 Effective Meaning Of Legacy Output Status

จาก master code เดิม ค่าที่ใช้บ่อยคือ:

| Value | Meaning in master business state | Slave electrical effect |
|---|---|---|
| `0` | none / no use | off |
| `1` | out | off |
| `2` | in | on |
| `3` | clean up | on |
| `4` | overtime | on |

ดังนั้น mapping ของ master รุ่นนี้คือ:

```text
status 1      => clear room bit
status 2..4   => set room bit
```

---

## 4. Master To PIC Display Protocol (AMP_V2D Generation)

อ้างอิงจาก:

- [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:2186)
- [dsp_v21/Main.c](/home/kaina/work/project/CCS/dsp_v21/Main.c:426)
- [DSPV2D.X/Main.c](/home/kaina/work/project/CCS/DSP_V2D/DSPV2D.X/Main.c:392)

### 4.1 Frame Shape

```text
':' 0xFE LED_INDEX LED_MODE
```

หรือใน display code ที่ใหม่ขึ้น:

```text
':' 0xFE PORT_REF LED_MODE [optional trailing byte in receive buffer variant]
```

สาระสำคัญคือ display board แยกด้วย address คงที่ `0xFE`

### 4.2 Field-By-Field

| Offset | Field | Size | Description |
|---|---|---|---|
| 0 | `':'` | 1 byte | start-of-frame |
| 1 | `0xFE` | 1 byte | display board marker |
| 2 | `LED_INDEX` หรือ `PORT_REF` | 1 byte | อ้างอิง card/port หรือ LED address |
| 3 | `LED_MODE` | 1 byte | pattern การแสดงผล |

### 4.3 LED Modes

จาก comment ใน master code:

| Value | Meaning |
|---|---|
| `0` | off |
| `1` | slow blink |
| `2` | fast blink |
| `3` | on |

อ้างอิงจาก [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:2186)

### 4.4 Display Mapping

ฝั่ง display จะตีความ byte ที่สามเป็น:

```text
CardNo = ref / 16
PortNo = ref % 16
PortCmd = mode & 0x0f
```

แล้วใช้ `LEDTABLE` map ไปยังตำแหน่งจริงใน LED matrix

อ้างอิงจาก:

- [dsp_v21/Main.c](/home/kaina/work/project/CCS/dsp_v21/Main.c:429)
- [DSP_V2D_1.X/main.c](/home/kaina/work/project/CCS/DSP_V2D/DSP_V2D_1.X/main.c:274)

### 4.5 How Blink Is Implemented

display board ไม่ได้กระพริบด้วย command แยกหลายครั้ง แต่ใช้ buffer สองชุด:

- `Dspbuf0`
- `Dspbuf1`

แล้ว timer interrupt สลับ mask เพื่อสร้าง slow/fast/on/off pattern

mapping โดยสรุป:

| Mode | Dspbuf0 | Dspbuf1 | Result |
|---|---|---|---|
| `0` | 0 | 0 | off |
| `1` | 0 | 1 | blink pattern A |
| `2` | 1 | 0 | blink pattern B |
| `3` | 1 | 1 | on |

อ้างอิงจาก:

- [dsp_v21/Main.c](/home/kaina/work/project/CCS/dsp_v21/Main.c:443)
- [DSPV2D.X/Main.c](/home/kaina/work/project/CCS/DSP_V2D/DSPV2D.X/Main.c:406)

### 4.6 Important AMP_V2D Behavior

ใน `AMP_V2D.C` การสั่ง LED ตอนเปลี่ยนสถานะห้องหลักเป็นแบบนี้:

- `IN` -> `UpdateLED(..., 3)`
- `CLEAN` -> `UpdateLED(..., 3)`
- `OT` -> `UpdateLED(..., 3)`
- `OUT` -> `UpdateLED(..., 0)`

ส่วน `slow blink` และ `fast blink` ถูกใช้เฉพาะช่วง warning จาก timer logic เช่นใกล้หมดเวลา cleanup/overtime ไม่ได้ map ตรงจาก business state หลักเสมอไป

อ้างอิงจาก:

- [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:520)
- [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:577)
- [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:606)
- [AMP_V2D.C](/home/kaina/work/project/CCS/AMP_V2D/AMP_V2D.C:609)

---

## 5. Legacy Master Internal State Model

อ้างอิงจาก:

- [AMP_V20/Main.C](/home/kaina/work/project/CCS/AMP_V20/Main.C:82)
- [AMP_V2_0/Main.c](/home/kaina/work/project/CCS/AMP_V2_0/Main.c:85)

ค่าที่พบใน master:

| Symbol | Value | Meaning |
|---|---|---|
| `STAOut` / `STAout` | `1` | guest out |
| `STAIn` / `STAin` | `2` | guest in |
| `STACup` / `STAcup` | `3` | clean up |
| `STAOt` / `STAot` | `4` | overtime |
| `STAot05` | `5` | OT 0.5 hr |
| `STAot10` | `6` | OT 1.0 hr |
| `STAot20` | `7` | OT 2.0 hr |

ข้อสำคัญ:

- ค่าพวกนี้เป็น `business state`
- ใน `AMP_V2D` ค่าพวกนี้ถูกยุบเป็น `bit on/off` สำหรับ slave board
- ส่วน display board ใช้ `0/3` เป็นหลัก และใช้ `1/2` เฉพาะ warning phase

---

## 6. Legacy Master Business Functions

ฟังก์ชันหลักที่พบใน master:

| Function | Legacy role |
|---|---|
| `SearchSlave` | หา room -> card/port |
| `SaveStatus` | เก็บสถานะห้องใน master RAM/XRAM |
| `GetStatus` | อ่านสถานะห้อง |
| `RoomOnoff` | ตัดสินใจ relay on/off ตาม status แล้วส่ง 485 |
| `UpdateLED` / `SendBlink` | คุม display board |
| `CheckOverTime` | ตรวจเวลาหมด / overtime |
| `CheckNormTime` | ตรวจเวลาปกติ / cleanup |
| `SaveAndPrint` / `PrnRecord` | บันทึกและพิมพ์รายงาน |

สรุปได้ว่า master เดิมทำทั้ง:

- state store
- rules engine
- scheduler/time engine
- reporting
- transport driver

---

## 7. Migration Mapping For New System

สำหรับระบบใหม่ ให้ถือว่า:

- PIC อยู่ในส่วน `slave` และ `display`
- ไม่มี PIC ที่ถือ business logic
- `backend` เป็น source of truth
- transport ปกติคือ `backend/local controller -> RS-485 -> PIC`
- `esp32-gateway` เป็นทางเลือกเสริม เมื่อ topology หน้างานไม่เหมาะกับการลาก RS-485 ตรง

### 7.1 Function Migration Table

| Legacy function / concern | New home | Reason |
|---|---|---|
| `SearchSlave` | `backend` | mapping ห้องเป็น config/business data |
| `SaveStatus` | `backend` | canonical room state ต้องอยู่ backend |
| `GetStatus` | `backend` | dashboard/handheld อ่านจาก backend |
| `GetStatusTime` | `backend` | runtime/check-in time เป็น business record |
| `SaveOtime` / `GetOtime` | `backend` | overtime policy เป็น business logic |
| `CheckOverTime` | `backend` | rule engine / scheduler |
| `CheckNormTime` | `backend` | workflow timing |
| `SaveAndPrint` / `PrnRecord` / report | `backend` | report, audit, export อยู่ server |
| `RoomOnoff` business decision | `backend` | ตัดสินว่าควร on/off จาก state ไหน |
| `Out485_16` transport | `RS-485 adapter` หรือ `esp32-gateway` | ปกติส่งตรงจาก local controller; ถ้าสายไกลค่อยใช้ gateway |
| `Tx485` | `RS-485 adapter` หรือ `esp32-gateway` | UART/RS485 driver |
| `SendBlink` | `RS-485 adapter` หรือ `esp32-gateway` | serial protocol adapter ไป PIC display |
| `WakeupSlave` ถ้าเป็น init/refresh transport | `RS-485 adapter` หรือ `esp32-gateway` | transport-side task |
| keypad/LCD/password UI | `dashboard` หรือ `handheld` | ย้ายออกจาก embedded master |
| RTC-local timing | `backend` | ใช้ server time กลาง |
| physical output latch / `WritePORT` | `PIC slave firmware` | เป็น hardware responsibility |
| LED matrix scan / blink buffers | `PIC display firmware` | เป็น hardware responsibility |
| UART frame receive + parse on slave | `PIC slave firmware` | รับ protocol เก่า |
| UART frame receive + parse on display | `PIC display firmware` | รับ protocol เก่า |

### 7.2 New Division Of Responsibilities

#### Backend

- room mapping
- state machine
- check-in / check-out / extend
- cleaning workflow
- anti-fraud / audit
- overtime / timeout logic
- API + SSE + MQTT

#### Direct RS-485 Path

- local controller หรือ backend-side service เปิด serial/RS-485 ตรง
- ส่ง `legacy slave frame` ไป PIC slave
- ส่ง `legacy display frame` ไป PIC display
- เก็บ timeout / retry / delivery result

#### ESP32 Gateway (Optional)

- subscribe command จาก backend
- แปลง command เป็น `legacy slave frame`
- แปลง LED command เป็น `legacy display frame`
- จัดการ UART/RS485 timing
- เก็บ delivery result / timeout / retry
- publish ack/result กลับ backend

#### PIC Slave Firmware

- decode frame แบบ legacy
- validate address
- แปลง nibble status เป็น output on/off
- latch output 16 channels

#### PIC Display Firmware

- decode frame แบบ legacy display
- map card/port ไป matrix position
- render off / slow / fast / on

---

## 8. Recommended Compatibility Strategy

ถ้าจะใช้ PIC เดิมต่อในระบบใหม่ แนวทางที่เหมาะที่สุดคือ:

1. backend สั่งงานด้วย semantic command เช่น `check_in`, `open_room`, `start_cleaning`
2. backend แปลง semantic state เป็น desired room/device state
3. transport layer map state นั้นไปเป็น:
   - slave frame สำหรับ output board
   - display frame สำหรับ LED board
4. PIC slave/display ทำงานแบบเดิมให้มากที่สุด

ข้อดี:

- reuse firmware เดิมได้มาก
- แยก business logic ออกจาก embedded ชัดเจน
- backend กลายเป็น source of truth ตามสถาปัตยกรรมใหม่

---

## 9. Key Risks From Legacy Protocol

- slave firmware ดูเหมือนจะไม่ enforce checksum เต็มรูปแบบในบางเวอร์ชัน
- protocol เป็น write-oriented มาก และไม่มี rich ACK/telemetry แบบ modern bus
- business state หลายค่า collapse เป็นแค่ on/off ที่ฝั่ง slave
- display protocol เป็น one-way control protocol ไม่ได้ส่ง state กลับ

ดังนั้นถ้าจะใช้ต่อในระบบใหม่ ควรให้ transport layer ซึ่งอาจเป็น local RS-485 service หรือ `esp32-gateway` ทำหน้าที่เพิ่ม:

- timeout detection
- retry
- command correlation
- state reconciliation ระหว่าง desired state กับ last sent state

---

## 10. Action To Legacy Frame Mapping

ส่วนนี้สรุปว่า action ฝั่งใหม่ควรถูกแปลงไปเป็น `legacy slave frame` และ `legacy display frame` อย่างไร

หมายเหตุสำคัญ:

- การแปลงนี้เป็น `transport mapping`
- business rule ว่าควรเปิด/ปิดอุปกรณ์อะไร ตัดสินที่ backend ก่อน
- เลขจริงของ `ADDR`, `PORT_REF`, `card`, `port` มาจาก room mapping ใน backend

### 10.1 Symbols Used

| Symbol | Meaning |
|---|---|
| `ADDR` | address ของ PIC slave board |
| `card` | หมายเลข card/slave ของห้อง |
| `port` | output index 0-15 บน card |
| `PORT_REF` | `(card * 16) + port` สำหรับ display board |
| `S` | legacy room status nibble |
| `CHKSUM` | two's complement checksum ของ slave frame |

### 10.2 Legacy Status Values Used For Transport

| Semantic state | `S` value | Relay effect | LED suggestion |
|---|---|---|---|
| `vacant` / `check_out` | `1` | off | off |
| `occupied` / `check_in` | `2` | on | on |
| `cleaning` / `start_cleaning` | `3` | on | on |
| `overstay` / `extend timeout` | `4` | on | on |

### 10.3 Slave Frame Builder Rule

สำหรับห้องหนึ่งห้อง:

1. หา `ADDR` และ `port`
2. เริ่มจาก relay bitmask 16 บิตของ card ใบนั้น
3. ถ้า `S >= 2` ให้ set bit ของ `port`
4. ถ้า `S < 2` ให้ clear bit ของ `port`
5. bit อื่นคงค่าเดิมของ card นั้นไว้
5. สร้าง frame:

```text
':' ADDR DATA_LO DATA_HI CHKSUM
```

โดย:

- `DATA_LO = bitmask & 0x00ff`
- `DATA_HI = (bitmask >> 8) & 0x00ff`

### 10.4 Display Frame Builder Rule

สำหรับห้องหนึ่งห้อง:

```text
PORT_REF = (card * 16) + port
```

แล้วส่ง:

```text
':' 0xFE PORT_REF LED_MODE
```

โดย `LED_MODE` ที่แนะนำ:

| Semantic state | LED_MODE |
|---|---|
| `vacant` | `0` |
| `occupied` | `3` |
| `cleaning` | `3` |
| `overstay` | `3` |

### 10.5 Action Mapping Table

| New action | Backend semantic result | Slave status `S` | Display mode | Notes |
|---|---|---|---|---|
| `check_in` | room becomes `occupied` | `2` | `3` | เปิด room power |
| `check_out` | room becomes `vacant` | `1` | `0` | ปิด room power |
| `extend_stay` | keep `occupied` | `2` | `3` | ถ้าไม่เกินเวลาให้คง occupied |
| `open_room` | pulse or temporary unlock event | no change by default | optional blink | ถ้าเปิดแค่ door strike ไม่ควร rewrite room state ถาวร |
| `close_room` | no state change or ensure secure output off | depends on hardware mapping | optional off | ใช้เฉพาะถ้ามี dedicated output |
| `start_cleaning` | room becomes `cleaning` | `3` | `3` | `AMP_V2D` เปิด LED ค้างตอนเริ่ม clean |
| `finish_cleaning` | room becomes `vacant` | `1` | `0` | กลับห้องว่าง |
| `set_makeup` | waiting_for_cleaning | keep previous relay state | `1` or custom | legacy protocol ไม่มีสถานะ makeup ตรง ๆ |
| `set_urgent_makeup` | urgent cleaning request | keep previous relay state | `2` | ใช้ display สื่อ urgency |
| `set_dnd` | do not disturb flag | keep previous relay state | custom or unchanged | legacy relay frame ไม่มี DND ตรง ๆ |
| `clear_dnd` | clear DND flag | keep previous relay state | custom or unchanged | ถ้าไม่มี LED เฉพาะก็ไม่ต้องส่ง |

### 10.6 Recommended Compatibility Rules

เพื่อให้เข้ากับ PIC เดิมโดยเปลี่ยนน้อยที่สุด:

- `check_in` -> ใช้ `S=2`
- `check_out` -> ใช้ `S=1`
- `start_cleaning` -> ใช้ `S=3`
- `finish_cleaning` -> ใช้ `S=1`
- `overstay` ภายใน backend -> ใช้ `S=4`

ส่วน action ที่ legacy protocol ไม่มีความหมายตรง เช่น `DND`, `makeup`, `urgent_makeup`, `open_room` ควรจัดการแบบนี้:

- เก็บ state จริงที่ backend
- ถ้ามี dedicated output/LED ในหน้างาน ค่อยเพิ่ม transport mapping เฉพาะ site
- ถ้ายังไม่มี ให้ใช้ display pattern เท่าที่สื่อได้ หรือไม่ส่งลง PIC เลย

### 10.7 Example

สมมติ:

- room `1205`
- mapped to `ADDR=0x03`
- `card=3`
- `port=5`
- action = `start_cleaning`

ผลที่ transport ควรสร้าง:

- relay bit ของ port 5 = `1`
- display frame = `':' 0xFE 0x35 0x03`

โดย `0x35` มาจาก:

```text
PORT_REF = (3 * 16) + 5 = 53 = 0x35
```

bitmask และ frame ที่ได้:

```text
bitmask  = 0x0020
DATA_LO  = 0x20
DATA_HI  = 0x00
CHKSUM   = two's complement of ':' + 0x03 + 0x20 + 0x00 = 0xA3

slave   = 3A 03 20 00 A3
display = 3A FE 35 03
```

### 10.8 Golden Examples For AMP_V2D

ตัวอย่างนี้ใช้ room mapping เดียวกัน:

- room `1205`
- `ADDR=0x03`
- `card=3`
- `port=5`

#### `check_in`

```text
slave   = 3A 03 20 00 A3
display = 3A FE 35 03
```

#### `check_out`

```text
slave   = 3A 03 00 00 C3
display = 3A FE 35 00
```

#### `start_cleaning`

```text
slave   = 3A 03 20 00 A3
display = 3A FE 35 03
```

#### `overstay`

```text
slave   = 3A 03 20 00 A3
display = 3A FE 35 03
```

#### `extend_stay`

```text
slave   = 3A 03 20 00 A3
display = 3A FE 35 03
```
