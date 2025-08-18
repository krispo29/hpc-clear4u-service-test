# การเปลี่ยนแปลง Workflow ของ MAWB

เอกสารนี้สรุปการเปลี่ยนแปลงที่เกิดขึ้นกับระบบ Backend API เพื่อรองรับ Workflow การจัดการ MAWB ตามที่ระบุใน `howto-th.md` ข้อ 1.1.1

## การเปลี่ยนแปลงที่เกิดขึ้น

### 1. Endpoint ใหม่: `POST /api/v1/mawbinfo/{uuid}/send-to-customer`

ได้เพิ่ม Endpoint ใหม่เพื่อรองรับการส่ง MAWB ไปให้ลูกค้ายืนยัน โดยมีการเปลี่ยนแปลงในไฟล์ต่างๆ ดังนี้:

- **`server/mawbinfo.go`**:
    - เพิ่ม `sendToCustomer` handler function เพื่อจัดการ request ที่เข้ามา.
    - เพิ่ม route `POST /send-to-customer` ภายใต้ `/{uuid}`.

- **`outbound/mawbinfo/service.go`**:
    - เพิ่ม `SendToCustomer` function ใน `Service` interface และ implement logic การทำงาน.
    - Logic จะตรวจสอบสถานะปัจจุบันของ MAWB หากเป็น `Draft` หรือ `RejectedByCustomer` จะเปลี่ยนสถานะเป็น `Pending`.

- **`outbound/mawbinfo/repository.go`**:
    - เพิ่ม `UpdateMawbInfoStatus` function สำหรับอัปเดตสถานะของ MAWB ในฐานข้อมูล.
    - เพิ่ม `GetMawbInfoStatus` function สำหรับดึงสถานะปัจจุบันของ MAWB.
    - แก้ไข `createTableIfNotExists` เพื่อเพิ่ม column `status` ใน `tbl_mawb_info`.
    - แก้ไข function `CreateMawbInfo`, `GetMawbInfo`, `GetAllMawbInfo` ให้รองรับ field `status`.

- **`outbound/mawbinfo/mawbinfo.go`**:
    - เพิ่ม `Status` field ใน `MawbInfoResponse` struct.
    - เพิ่ม constants สำหรับสถานะต่างๆ ของ MAWB (เช่น `StatusDraft`, `StatusPending`).

## สรุปการทำงาน

เมื่อ Admin เรียกใช้งาน Endpoint `POST /api/v1/mawbinfo/{uuid}/send-to-customer` ระบบจะ:
1. ตรวจสอบสถานะของ MAWB ที่ระบุด้วย `{uuid}`.
2. หากสถานะเป็น `Draft` หรือ `RejectedByCustomer` ระบบจะอัปเดตสถานะเป็น `Pending`.
3. หากสถานะไม่ถูกต้อง ระบบจะ trả về error.

การเปลี่ยนแปลงนี้เป็นส่วนหนึ่งของการสร้าง Workflow การอนุมัติ MAWB ที่สมบูรณ์ต่อไป.