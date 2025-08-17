# วิธีทำ: การสร้าง Workflow สำหรับ MAWB และ Cargo Manifest

เอกสารนี้สรุปแนวทางการพัฒนา Frontend และ Backend สำหรับ Workflow การอนุมัติ MAWB (Master Air Waybill) และ Cargo Manifest ตามที่ระบุในข้อ 5.4-5.8

## 1. การพัฒนาระบบหลังบ้าน (Backend API)

ฝั่ง Backend ต้องเตรียม API Endpoints หลายตัวเพื่อจัดการสถานะของ MAWB และ Cargo Manifest เราจะสมมติว่ามี Model `MAWBInfo` ในฐานข้อมูลซึ่งมีฟิลด์ `status` (เช่น 'Draft', 'Pending', 'ConfirmedByCustomer', 'RejectedByCustomer', 'ConfirmedByAdmin', 'Canceled') และฟิลด์ `cargoManifestStatus` (เช่น 'Draft', 'Confirmed', 'Rejected')

### 1.1 Endpoints สำหรับสถานะ MAWB

Endpoints เหล่านี้จะจัดการวงจรชีวิต (Lifecycle) ของเอกสาร MAWB โดย `{uuid}` ใน URL คือรหัสเฉพาะของ MAWB นั้นๆ

#### 1.1.1. การส่ง MAWB ให้ลูกค้ายืนยัน

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/send-to-customer`
*   **คำอธิบาย:** สำหรับให้ Admin ใช้ส่งร่าง MAWB ไปให้ลูกค้ายืนยัน
*   **ผู้ใช้งาน:** Admin
*   **การทำงาน (Logic):**
    1.  ตรวจสอบสิทธิ์ว่าเป็น Admin
    2.  ค้นหา MAWB ด้วย `{uuid}`
    3.  ตรวจสอบว่าสถานะปัจจุบันเป็น 'Draft' หรือ 'RejectedByCustomer'
    4.  อัปเดตสถานะ MAWB เป็น `'Pending'`
    5.  (ทางเลือก) ส่งอีเมลแจ้งเตือนไปยังลูกค้า
*   **Success Response:** `200 OK` พร้อมข้อความสำเร็จ
*   **Error Response:** `403 Forbidden` (ถ้าไม่ใช่ Admin), `404 Not Found` (ถ้าไม่พบ MAWB), `409 Conflict` (ถ้าสถานะไม่ถูกต้อง)

#### 1.1.2. การยืนยัน MAWB (โดยลูกค้า)

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/customer-confirm`
*   **คำอธิบาย:** สำหรับให้ลูกค้าใช้ยืนยันว่าร่าง MAWB ถูกต้อง
*   **ผู้ใช้งาน:** Customer
*   **การทำงาน (Logic):**
    1.  ตรวจสอบว่าเป็นลูกค้าและเป็นเจ้าของ MAWB
    2.  ค้นหา MAWB
    3.  ตรวจสอบว่าสถานะปัจจุบันเป็น `'Pending'`
    4.  อัปเดตสถานะเป็น `'Confirmed'`
    5.  (ทางเลือก) ส่งอีเมลแจ้งเตือนไปหา Admin
*   **Success Response:** `200 OK`
*   **Error Response:** `403 Forbidden`, `404 Not Found`, `409 Conflict`

#### 1.1.3. การปฏิเสธ MAWB (โดยลูกค้า)

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/customer-reject`
*   **คำอธิบาย:** สำหรับให้ลูกค้าใช้ปฏิเสธร่าง MAWB
*   **ผู้ใช้งาน:** Customer
*   **การทำงาน (Logic):**
    1.  ตรวจสอบสิทธิ์และความเป็นเจ้าของ
    2.  ค้นหา MAWB
    3.  ตรวจสอบว่าสถานะปัจจุบันเป็น `'Pending'`
    4.  อัปเดตสถานะเป็น `'Rejected'`
    5.  (ทางเลือก) บันทึกเหตุผลการปฏิเสธจาก request body
    6.  (ทางเลือก) ส่งอีเมลแจ้งเตือนไปหา Admin
*   **Success Response:** `200 OK`
*   **Error Response:** `403 Forbidden`, `404 Not Found`, `409 Conflict`

#### 1.1.4. การยืนยัน MAWB (โดย Admin)

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/confirm`
*   **คำอธิบาย:** สำหรับให้ Admin ยืนยันขั้นสุดท้าย (หลังจากลูกค้าสร้างเอง หรือหลังจากลูกค้ายืนยันฉบับที่ Admin สร้าง)
*   **ผู้ใช้งาน:** Admin
*   **การทำงาน (Logic):**
    1.  ตรวจสอบว่าเป็น Admin
    2.  ค้นหา MAWB
    3.  ตรวจสอบว่าสถานะเป็น `'Confirmed'` (ลูกค้ายืนยันแล้ว) หรือ `'Draft'` (ลูกค้าสร้างเอง)
    4.  อัปเดตสถานะเป็น `'ConfirmedByAdmin'`
*   **Success Response:** `200 OK`
*   **Error Response:** `403 Forbidden`, `404 Not Found`, `409 Conflict`

#### 1.1.5. การปฏิเสธ MAWB (โดย Admin)

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/reject`
*   **คำอธิบาย:** สำหรับให้ Admin ปฏิเสธร่าง MAWB ที่ลูกค้าเป็นคนสร้าง
*   **ผู้ใช้งาน:** Admin
*   **การทำงาน (Logic):**
    1.  ตรวจสอบว่าเป็น Admin
    2.  ค้นหา MAWB
    3.  ตรวจสอบว่าสถานะเป็น `'Draft'` และ `createdBy` คือ 'customer'
    4.  อัปเดตสถานะเป็น `'Rejected'`
*   **Success Response:** `200 OK`
*   **Error Response:** `403 Forbidden`, `404 Not Found`, `409 Conflict`

#### 1.1.6. การยกเลิก MAWB (โดย Admin)

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/cancel`
*   **คำอธิบาย:** สำหรับให้ Admin ยกเลิกกระบวนการ MAWB (เช่น หลังจากลูกค้าปฏิเสธ)
*   **ผู้ใช้งาน:** Admin
*   **การทำงาน (Logic):**
    1.  ตรวจสอบว่าเป็น Admin
    2.  ค้นหา MAWB
    3.  ตรวจสอบว่าสถานะเป็น `'Rejected'`
    4.  อัปเดตสถานะเป็น `'Canceled'`
*   **Success Response:** `200 OK`
*   **Error Response:** `403 Forbidden`, `404 Not Found`, `409 Conflict`

### 1.2 Endpoints สำหรับ Cargo Manifest

Endpoints ส่วนนี้จะจัดการสถานะของ Cargo Manifest ซึ่งตรงไปตรงมามากกว่า

#### 1.2.1. การดู/พรีวิว Cargo Manifest

*   **Endpoint:** `GET /api/v1/mawbinfo/cargo-manifest/print/{uuid}`
*   **คำอธิบาย:** สร้างและส่งคืนไฟล์ PDF ของร่าง Cargo Manifest
*   **ผู้ใช้งาน:** Admin/Customer
*   **การทำงาน (Logic):**
    1.  ค้นหาข้อมูล Cargo Manifest ที่เชื่อมกับ MAWB `{uuid}`
    2.  สร้างเอกสาร PDF จากข้อมูล
    3.  ส่งคืนไฟล์ PDF ใน response พร้อม `Content-Type: application/pdf`
*   **Success Response:** `200 OK` พร้อมข้อมูล PDF
*   **Error Response:** `404 Not Found`

#### 1.2.2. การยืนยัน Cargo Manifest

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/cargo-manifest/confirm`
*   **คำอธิบาย:** ยืนยันร่าง Cargo Manifest
*   **ผู้ใช้งาน:** Admin/Customer (ขึ้นอยู่กับ Logic ของธุรกิจ)
*   **การทำงาน (Logic):**
    1.  ค้นหา MAWB
    2.  ตรวจสอบว่า `cargoManifestStatus` เป็น `'Draft'`
    3.  อัปเดต `cargoManifestStatus` เป็น `'Confirmed'`
*   **Success Response:** `200 OK`
*   **Error Response:** `404 Not Found`, `409 Conflict`

#### 1.2.3. การปฏิเสธ Cargo Manifest

*   **Endpoint:** `POST /api/v1/mawbinfo/{uuid}/cargo-manifest/reject`
*   **คำอธิบาย:** ปฏิเสธร่าง Cargo Manifest
*   **ผู้ใช้งาน:** Admin/Customer
*   **การทำงาน (Logic):**
    1.  ค้นหา MAWB
    2.  ตรวจสอบว่า `cargoManifestStatus` เป็น `'Draft'`
    3.  อัปเดต `cargoManifestStatus` เป็น `'Rejected'`
*   **Success Response:** `200 OK`
*   **Error Response:** `404 Not Found`, `409 Conflict`

---

## 2. การพัฒนาระบบหน้าบ้าน (Frontend)

ฝั่ง Frontend คือ React Component (`EditMAWBInfo`) ที่แสดงข้อมูล MAWB และปุ่มดำเนินการต่างๆ ตามบทบาทของผู้ใช้และสถานะของเอกสาร

### 2.1 การจัดการ State

Component ต้องมีการจัดการ State ดังนี้:

```javascript
const [formData, setFormData] = useState({
  // ... ฟิลด์อื่นๆ ของฟอร์ม
  status: "Draft", // เช่น Draft, Pending, Confirmed, Rejected
  cargoManifestStatus: "Draft", // เช่น Draft, Confirmed, Rejected
  createdBy: "admin", // 'admin' หรือ 'customer'
});
const [userRole, setUserRole] = useState("admin"); // 'admin' หรือ 'customer'
const [isSubmitting, setIsSubmitting] = useState(false);
```

-   `formData`: เก็บข้อมูลทั้งหมดของ MAWB รวมถึงสถานะต่างๆ ซึ่งควรดึงมาจาก Backend ผ่าน `useEffect`
-   `userRole`: บทบาทของผู้ใช้ที่ล็อกอินอยู่ อาจเก็บใน Local Storage หรือ Context
-   `isSubmitting`: ตัวแปร Boolean เพื่อปิดการใช้งานปุ่มระหว่างเรียก API ป้องกันการกดซ้ำ

### 2.2 ฟังก์ชันสำหรับเรียก API (API Handlers)

สร้างฟังก์ชันสำหรับเรียก API ของ Backend โดยแต่ละฟังก์ชันจะจัดการการเรียก API, แสดงข้อความแจ้งเตือน (success/error), และอัปเดต State ในหน้าเว็บ

```javascript
// ตัวอย่างฟังก์ชัน "Send to Customer"
const handleSendToCustomer = async () => {
  setIsSubmitting(true);
  try {
    await axiosInstance.post(`/v1/mawbinfo/${uuid}/send-to-customer`);
    // อัปเดต State ทันทีเพื่อให้หน้าเว็บเปลี่ยนแปลง
    setFormData((prev) => ({ ...prev, status: "Pending" }));
    toast.success("ส่งให้ลูกค้าเรียบร้อยแล้ว!");
  } catch (error) {
    toast.error("ไม่สามารถส่งให้ลูกค้าได้");
  } finally {
    setIsSubmitting(false);
  }
};

// ควรสร้างฟังก์ชันแบบเดียวกันสำหรับทุก Action:
// - handleConfirmMAWB (by admin)
// - handleRejectMAWB (by admin)
// - handleCancelMAWB (by admin)
// - handleCustomerConfirmMAWB
// - handleCustomerRejectMAWB
// - handleConfirmCargoManifest
// - handleRejectCargoManifest
// - handleViewDraftCargoManifest
```

### 2.3 การแสดงผล UI ตามเงื่อนไข (Conditional Rendering)

หัวใจหลักของ Frontend คือการแสดงปุ่มให้ถูกต้องตามเงื่อนไขของ `userRole`, `formData.status`, `formData.createdBy`, และ `formData.cargoManifestStatus`

#### Logic ของปุ่ม MAWB:

```jsx
<div className="action-buttons">
  {/* กรณี 5.4: Admin เป็นคนสร้าง */}
  {userRole === "admin" && formData.createdBy === "admin" && (
    <>
      {formData.status === "Draft" && (
        <Button onClick={handleSendToCustomer} disabled={isSubmitting}>
          Send to Customer
        </Button>
      )}
      {/* สมมติว่าเมื่อลูกค้ายืนยัน สถานะจะเปลี่ยนเป็น 'Confirmed' */}
      {formData.status === "Confirmed" && (
        <Button onClick={handleConfirmMAWB} disabled={isSubmitting}>
          Confirm Draft MAWB (Admin)
        </Button>
      )}
      {formData.status === "Rejected" && (
        <>
          <Button onClick={handleSendToCustomer} disabled={isSubmitting}>
            Resend to Customer
          </Button>
          <Button variant="danger" onClick={handleCancelMAWB} disabled={isSubmitting}>
            Cancel MAWB
          </Button>
        </>
      )}
    </>
  )}

  {/* กรณี 5.5: ลูกค้าเป็นคนสร้าง และ Admin เข้ามาดู */}
  {userRole === "admin" && formData.createdBy === "customer" && (
    <>
      {formData.status === "Draft" && (
        <>
          <Button onClick={handleConfirmMAWB} disabled={isSubmitting}>
            Confirm Draft MAWB
          </Button>
          <Button variant="danger" onClick={handleRejectMAWB} disabled={isSubmitting}>
            Reject Draft MAWB
          </Button>
        </>
      )}
    </>
  )}

  {/* กรณีสำหรับฝั่งลูกค้า */}
  {userRole === "customer" && formData.status === "Pending" && (
    <>
        <Button onClick={handleCustomerConfirmMAWB} disabled={isSubmitting}>
            Confirm MAWB
        </Button>
        <Button variant="danger" onClick={handleCustomerRejectMAWB} disabled={isSubmitting}>
            Reject MAWB
        </Button>
    </>
  )}
</div>
```

#### Logic ของปุ่ม Cargo Manifest (5.6, 5.7, 5.8):

ส่วนนี้ควรจะอยู่ใน Component "Cargo Manifest Details"

```jsx
{cargoManifest && (
  <div className="d-flex justify-content-end gap-2">
    {/* 5.6: ปุ่มดูเอกสาร */}
    <Button variant="info" onClick={handleViewDraftCargoManifest}>
      <FaPrint /> View Draft Cargo Manifest
    </Button>

    {/* 5.7 & 5.8: ปุ่มยืนยัน/ปฏิเสธ */}
    {formData.cargoManifestStatus === "Draft" && (
      <>
        <Button variant="success" onClick={handleConfirmCargoManifest} disabled={isSubmitting}>
          <FaCheck /> Confirm Cargo Manifest
        </Button>
        <Button variant="danger" onClick={handleRejectCargoManifest} disabled={isSubmitting}>
          <FaTimes /> Reject Cargo Manifest
        </Button>
      </>
    )}
  </div>
)}
```

นี่คือแนวทางทั้งหมดสำหรับการพัฒนาฟีเจอร์ที่ต้องการ สิ่งสำคัญคือการจับคู่ State ของแอปพลิเคชัน (`status`, `userRole`) ให้เข้ากับ Logic ทางธุรกิจที่กำหนดไว้ใน Requirement
