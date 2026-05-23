# CCS2 Legacy Master Manual Summary

สรุปจาก [คู่มือ ccs2.pdf](/home/kaina/work/project/ccs2_plus/คู่มือ%20ccs2.pdf) แบบหน้า-by-หน้า เพื่อจับข้อมูลที่ `pdftotext` อ่านไม่ครบ

---

## Overall Takeaways

- master เดิมเป็นอุปกรณ์หน้า front office พร้อม keypad + LCD
- ระบบประกอบด้วย `master`, `relay cabinet`, `display board`
- operator ใช้งานผ่าน flow `check-in`, `check-out`, `temporary`, `extend`
- มีระบบพิมพ์ report 4 แบบ
- มี wiring และ installation guide ที่ยืนยันว่า relay board และ display board แยกหน้าที่ชัดเจน

---

## Page-By-Page Summary

### Page 1

- หน้าปกคู่มือการใช้งาน
- ชื่อระบบ `ระบบเช็คอิน/เช็คเอาท์`
- ภาพรวมอุปกรณ์หลักคือ master, relay cabinet, display board

### Page 2

- หน้า `ส่วนประกอบ`
- ระบุ 3 องค์ประกอบหลัก:
  - ชุด master 1 เครื่อง
  - ตู้ควบคุมรีเลย์
  - แผงไฟโชว์แสดงสถานะของห้องพัก 1 แผง

### Page 3

- สารบัญ
- หัวข้อสำคัญที่คู่มือครอบคลุม:
  - ส่วนประกอบของระบบ
  - หลักการทำงานของระบบ CCS
  - การตั้งเวลาของเครื่อง
  - การดูสถานะห้องพัก
  - การเปิด-ปิดไฟภายในห้องพัก
  - การพิมพ์ข้อมูลแบบ `PERIOD`, `RECORD`, `SUMMARY`
  - การติดตั้งผ่าน printer
  - การติดตั้งผ่านตู้ควบคุมรีเลย์
  - การเดินสายไฟห้องพัก
  - ปัญหาที่พบบ่อยและการแก้ไขเบื้องต้น

### Page 4

- อธิบายหลักการทำงานของ `CCS FRONT OFFICE CHECK-IN CHECK-OUT SYSTEM`
- ยืนยันว่า front office เป็นศูนย์กลางของการควบคุมไฟห้องพัก
- ทุกครั้งที่มี `check-in/check-out` เครื่องจะเก็บข้อมูลวัน/เดือน/ปี/เบอร์ห้อง/เวลา
- เครื่องเก็บข้อมูลได้แม้ไฟดับ
- master 1 ชุดควบคุมได้จำนวนห้องมาก
- ระบบถูกออกแบบให้ใช้งานง่ายและตรวจสอบกับคอมพิวเตอร์ได้

### Page 5

- ขั้นตอนการตั้งเวลาเครื่อง
- ใช้เมนู `Set date/time`
- มี password ก่อนเข้าเมนูตั้งค่า
- master เดิมมี operator setup flow ในตัวเครื่องเอง

### Page 6

- การดูสถานะของห้องพัก
- คู่มือบอกชัดว่ามีอย่างน้อย 2 สถานะ:
  - หมายเลขห้องไม่กระพริบ = ห้องว่าง
  - หมายเลขห้องกระพริบ = มีลูกค้าพักอยู่
- แสดงว่าหนึ่งในบทบาทสำคัญของ display board คือสื่อ occupied/vacant ให้ operator เห็นรวดเร็ว

### Page 7

- การเปิด-ปิดไฟในห้องพัก
- มี flow หลัก 4 แบบ:
  - `เช็คอิน`
  - `เช็คเอาท์`
  - `ชั่วคราว`
  - `ต่อเวลา`
- คู่มือบอกชัดว่า operator ต้องเลือกห้องก่อน แล้วค่อยเลือก action
- `ต่อเวลา` ใช้ตัวเลข 1, 2, 3 เพื่อแทนชั่วโมงหรือครึ่งชั่วโมงบางแบบ

### Page 8

- เมนูการพิมพ์ข้อมูล
- ยืนยัน report 4 แบบ:
  - `PERIAD/PERIOD`
  - `RECORD`
  - `SUMMARY`
  - `INFO`
- ต้องใส่ password ก่อนเข้าหน้าพิมพ์

### Page 9

- อธิบายการพิมพ์ `RECORD`
- มีการเลือกช่วงวันเริ่มต้นและวันสิ้นสุด
- แสดงว่า master เดิมรองรับ date-range query สำหรับ report

### Page 10

- อธิบายการพิมพ์ `SUMMARY`
- มีการเลือกช่วงวันเริ่มต้นและวันสิ้นสุดเช่นเดียวกับ `RECORD`

### Page 11

- ตัวอย่างการพิมพ์ `PERIOD`
- field ที่เห็นชัด:
  - `ROOM`
  - `GUEST IN`
  - `GUEST OUT`
  - `STATUS`
  - `REMARK`
- annotation ในเอกสารบอกความหมายของเครื่องหมาย เช่น
  - `*` = ค้างคืน
  - `**` = ชั่วคราว
- นี่เป็นหลักฐานสำคัญว่าระบบเดิมมี `stay mode / case` แฝงอยู่ใน report

### Page 12

- ตัวอย่างการพิมพ์ `RECORD`
- field ที่เห็นชัด:
  - `DATE`
  - `ROOM`
  - `GUEST IN`
  - `GUEST OUT`
  - `CLEAN UP`
  - `CASE`
  - `REMARK`
- ยืนยันว่าระบบเดิมแยก `clean up` เป็นข้อมูลอีกแกนหนึ่ง ไม่ใช่แค่ occupied/vacant

### Page 13

- ตัวอย่างการพิมพ์ `SUMMARY`
- field ที่เห็นชัด:
  - `DATE`
  - `ROOM`
  - `GUEST IN`
  - `GUEST OUT`
  - `USED`
  - `CASE`
  - `REMARK`
- ใช้สรุปเวลาที่ใช้ห้องพักรายวัน

### Page 14

- ภาพ `การติดตั้ง`
- แสดงการเชื่อมระหว่าง:
  - ตู้เมนเบรกเกอร์
  - ตู้รีเลย์ในตู้
  - master
  - display board
  - เครื่องพิมพ์
- คู่มือระบุการใช้ `สายโทรศัพท์ขนาด 4 แกน (TIEV 4x0.65 sq.mm)`
- หน้านี้ยืนยัน physical topology แบบรวมศูนย์ที่ front office

### Page 15

- ภาพ `การติดตั้ง` ในมุมของห้องพัก
- เห็นกล่องควบคุมรีเลย์ต่อไปยังห้อง 101, 102, 110
- มีเส้นสาย 2 แกนไปยังห้อง
- มีเส้นสาย 4 แกนจาก master ไป display
- สะท้อนว่า room-side controller เป็น distributed boxes แต่ master เป็นศูนย์กลาง

### Page 16

- หน้า `การเดินสาย`
- แสดง topology แบบหลายบล็อก/หลายตู้รีเลย์
- ยืนยันว่าระบบเดิมรองรับการขยายหลายโซนหรือหลายบล็อกอาคาร
- มีสาย 4 แกนจาก master ไป display และสายสื่อสาร/ควบคุมไปตู้รีเลย์

### Page 17

- หน้าเกือบว่าง ไม่มีข้อมูลใช้งานสำคัญที่มองเห็น

---

## Important Design Implications For CCS-2PLUS PRO

- `check_in`, `check_out`, `temporary`, `extend` ควรยังเป็น semantic actions หลักใน backend ใหม่
- `stay mode` เช่น `overnight` และ `temporary` ควรถูกเก็บเป็น field จริงในระบบใหม่ ไม่ควรทิ้ง
- report model ใหม่ควรเก็บอย่างน้อย:
  - room
  - guest_in
  - guest_out
  - clean_up
  - used_duration
  - case
  - remark
- display board ในระบบเดิมมีความสำคัญต่อ operator workflow จึงควรยังคง layer นี้ไว้ถ้าหน้างานยังใช้
- relay layer กับ display layer เป็นคนละบทบาท ต้องไม่รวมกันโดยไม่ตั้งใจใน transport design
