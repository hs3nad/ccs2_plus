# CCS-2PLUS Board Port Plan

เอกสารนี้เก็บ port plan สำหรับ deployment ปัจจุบันที่ใช้ผ่าน `frpc`

| Board | Dashboard | SSH |
|---|---:|---:|
| Lyra | 6010 | 6110 |
| Pi1 | 6010 | 6110 |
| CCS-2PLUS dev | 6010 | 6110 |

หมายเหตุ:

- `Dashboard` คือ remote port ที่เปิดหน้า dashboard/backend ผ่าน `frpc`
- `SSH` คือ remote port สำหรับเข้าดูแลเครื่องผ่าน `frpc`
- dashboard/backend ปัจจุบันใช้ `REST + SSE` ไม่ใช้ `/ws`
- deployment ปัจจุบันของโปรเจกต์นี้ใช้ endpoint กลาง `34.87.151.202:6010` และ `34.87.151.202:6110`
