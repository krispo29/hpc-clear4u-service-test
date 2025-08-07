# Code Review: แพ็กเกจ `mawbinfo`

เอกสารนี้เป็นการรีวิวโค้ดของแพ็กเกจ `mawbinfo` ซึ่งเป็นส่วนหนึ่งของโปรเจกต์ `hpc-express-service` โดยจะสรุปภาพรวมของสถาปัตยกรรม, ฟังก์ชันการทำงาน, จุดแข็ง และข้อเสนอแนะในการปรับปรุงโค้ด

---

## 1. ภาพรวมและสถาปัตยกรรม (Overall Architecture)

แพ็กเกจ `mawbinfo` ได้รับการออกแบบมาอย่างดี โดยมีการแบ่ง Layer ของแอปพลิเคชันตามหลัก Clean Architecture ซึ่งช่วยให้โค้ดเป็นระเบียบ, ง่ายต่อการทำความเข้าใจ, บำรุงรักษา และทดสอบในอนาคต โครงสร้างหลักประกอบด้วย:

- **`mawbinfo/mawbinfo.go` (Models):** กำหนด Structs สำหรับ Request/Response (Data Transfer Objects) ทำให้เห็นโครงสร้างข้อมูลที่ใช้สื่อสารกันระหว่าง Layer ได้อย่างชัดเจน
- **`mawbinfo/service.go` (Business Logic):** จัดการตรรกะทางธุรกิจทั้งหมด เช่น การตรวจสอบข้อมูล (Validation), การคำนวณ, และการจัดการไฟล์แนบ
- **`mawbinfo/repository.go` (Data Access Layer):** รับผิดชอบการติดต่อกับฐานข้อมูล (PostgreSQL) ทั้งหมด แยกส่วนของการจัดการข้อมูลออกจาก Business Logic อย่างชัดเจน
- **`server/mawbinfo.go` (Delivery/Handler Layer):** ทำหน้าที่เป็นตัวกลางในการรับ HTTP Request, ส่งต่อไปยัง Service และจัดรูปแบบ Response กลับไปยัง Client โดยใช้ `chi` router

โครงสร้างนี้เป็นจุดแข็งที่สำคัญของโปรเจกต์ และเป็นแบบอย่างที่ดีในการพัฒนาซอฟต์แวร์

---

## 2. ฟังก์ชันการทำงาน (Functionality)

แพ็กเกจนี้จัดการข้อมูล "MAWB Info" ได้อย่างครบถ้วน (CRUD Operations) และมีความสามารถเพิ่มเติมที่สำคัญดังนี้:

- **Create:** สร้างข้อมูล MAWB Info ใหม่
- **Read:**
    - ดึงข้อมูลเฉพาะรายการ (by UUID)
    - ดึงข้อมูลทั้งหมด พร้อมความสามารถในการกรองข้อมูลตามช่วงวันที่ (start/end date)
- **Update:** อัปเดตข้อมูล MAWB Info ที่มีอยู่เดิม และรองรับการ **เพิ่ม** ไฟล์แนบใหม่เข้าไปในรายการเดิม
- **Delete:**
    - ลบข้อมูล MAWB Info ทั้งรายการ
    - ลบเฉพาะไฟล์แนบ (Attachment) ที่ต้องการ โดยมีการลบไฟล์ออกจาก File System และอัปเดตข้อมูลในฐานข้อมูลไปพร้อมกัน

---

## 3. จุดแข็ง (Strengths)

1.  **Separation of Concerns:** การแบ่ง Layer ชัดเจน ทำให้โค้ดแต่ละส่วนมีหน้าที่รับผิดชอบของตัวเอง ไม่ปะปนกัน
2.  **การใช้ Interfaces:** `Service` และ `Repository` ถูกกำหนดเป็น Interface ทำให้ง่ายต่อการทำ Dependency Injection และการเขียน Unit Test (โดยการสร้าง Mock)
3.  **การจัดการฐานข้อมูลที่ดี:**
    - `repository.go` มีฟังก์ชัน `createTableIfNotExists` ที่ช่วยสร้างตาราง `tbl_mawb_info` และเพิ่มคอลัมน์ `attachments` ให้อัตโนมัติ ทำให้การ Deploy หรือการเริ่มต้นใช้งานครั้งแรกทำได้สะดวก
    - การลบไฟล์แนบ (`DeleteMawbInfoAttachment`) มีการใช้ Transaction ของฐานข้อมูล ซึ่งช่วยรับประกันว่าข้อมูลจะถูกปรับปรุงอย่างถูกต้อง (Atomicity)
4.  **การตรวจสอบข้อมูล (Input Validation):**
    - มีการตรวจสอบข้อมูลที่รับเข้ามา (Request Body/Form) ทั้งใน Handler Layer (ใช้ `validator` library) และ Service Layer (ตรวจสอบค่าว่าง, รูปแบบวันที่, ค่าตัวเลข) ซึ่งเป็นการป้องกันข้อมูลที่ไม่ถูกต้องตั้งแต่เนิ่นๆ
5.  **การจัดการไฟล์แนบ:** การอัปเดตข้อมูล (Update) สามารถเพิ่มไฟล์แนบใหม่เข้าไปในรายการเดิมได้โดยไม่ลบของเก่า ซึ่งเป็นพฤติกรรมที่ถูกต้อง

---

## 4. จุดที่ควรปรับปรุง (Areas for Improvement)

1.  **การจัดการข้อผิดพลาด (Error Handling):**
    - **ความละเอียดของ Error:** ปัจจุบันมีการใช้ `errors.New(...)` หรือ `fmt.Errorf(...)` เป็นหลัก ซึ่งทำให้ Handler Layer ไม่สามารถแยกแยะประเภทของข้อผิดพลาดได้ (เช่น "Not Found" กับ "Invalid Input") และมักจะ trả về HTTP Status Code `400 Bad Request` เสมอ ควรพิจารณาสร้าง Custom Error Types (เช่น `ErrNotFound`, `ErrDuplicateRecord`) เพื่อให้ Handler สามารถ trả về status code ที่เหมาะสมกว่าได้ (เช่น `404 Not Found`)
    - **การจัดการข้อผิดพลาดในการลบไฟล์:** ใน `service.DeleteMawbInfoAttachment` หากลบไฟล์จาก File System (`os.Remove`) ไม่สำเร็จ โปรแกรมจะแค่ `fmt.Printf` ซึ่งข้อความนี้อาจจะหายไปใน Production ควรเปลี่ยนไปใช้ Logger ที่เหมาะสม และอาจต้องพิจารณานโยบายการจัดการในกรณีนี้ เช่น การลองใหม่ (Retry) หรือการสร้าง Background Job เพื่อตามลบไฟล์ในภายหลัง

2.  **การจัดการ Configuration:**
    - มีการ Hardcode ค่าบางอย่างไว้ในโค้ด เช่น `contextTimeout` (5 วินาที) หรือ `max-form-memory` (32MB) ใน `server/mawbinfo.go` ค่าเหล่านี้ควรถูกย้ายไปอยู่ในไฟล์ Configuration เพื่อให้ปรับเปลี่ยนได้ง่ายโดยไม่ต้องแก้ไขโค้ด

3.  **ความปลอดภัย (Security):**
    - **Path Traversal:** ใน `service.DeleteMawbInfoAttachment` มีความเสี่ยงที่ `fileName` อาจมี "path traversal characters" (เช่น `../../`) ซึ่งอาจทำให้ผู้ไม่หวังดีสามารถลบไฟล์อื่นๆ นอกเหนือจาก Directory ที่กำหนดได้ ควรมีการตรวจสอบและ Sanitize path ก่อนที่จะสั่งลบไฟล์
    - **SQL Injection:** แม้ว่าโค้ดส่วนใหญ่จะใช้ Prepared Statements ซึ่งปลอดภัย แต่ในฟังก์ชัน `GetAllMawbInfo` มีการต่อสตริง SQL โดยตรง (`" WHERE " + strings.Join(whereConditions, " AND ")`) แม้จะมีการตรวจสอบ format ของ `startDate` และ `endDate` แล้ว แต่แนวทางปฏิบัติที่ดีที่สุดคือการใช้ Placeholder (`?`) สำหรับทุกค่าที่มาจากผู้ใช้เสมอ เพื่อความปลอดภัยสูงสุด

4.  **โค้ดที่ซ้ำซ้อน (Code Duplication):**
    - **Validation Logic:** ฟังก์ชัน `validateInput` และ `validateUpdateInput` ใน `service.go` มีโค้ดที่เหมือนกันเกือบทั้งหมด สามารถรวมเป็นฟังก์ชันเดียวกันได้
    - **Database Scanning:** ตรรกะในการอ่านข้อมูลจากแถวของฐานข้อมูล (Scan) และการแปลง JSON ของ `attachments` เกิดขึ้นซ้ำๆ ในฟังก์ชัน `GetMawbInfo`, `GetAllMawbInfo`, และ `UpdateMawbInfo` สามารถแยกออกมาเป็นฟังก์ชัน Helper (เช่น `scanMawbInfoRow`) เพื่อลดความซ้ำซ้อนได้

---

## สรุป (Conclusion)

โดยรวมแล้ว แพ็กเกจ `mawbinfo` มีคุณภาพโค้ดที่ดี มีโครงสร้างที่แข็งแรงและเข้าใจง่าย จุดแข็งหลักคือการออกแบบสถาปัตยกรรมที่ชัดเจน หากมีการปรับปรุงในจุดที่แนะนำ (โดยเฉพาะด้าน Error Handling และ Security) จะทำให้โค้ดมีความสมบูรณ์และทนทานต่อข้อผิดพลาดมากยิ่งขึ้น
