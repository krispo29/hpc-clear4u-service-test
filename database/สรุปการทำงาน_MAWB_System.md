# สรุปการทำงาน MAWB System Integration - ภาษาไทย

## สิ่งที่ได้ทำไปแล้ว

### 1. สร้างระบบ Migration สำหรับฐานข้อมูล

#### ไฟล์ที่สร้างขึ้น:
- `database/migrations.go` - ตัวจัดการ migration หลัก
- `cmd/migrate/main.go` - โปรแกรมสำหรับรัน migration
- `cmd/verify-schema/main.go` - โปรแกรมตรวจสอบ schema
- `cmd/clear-migrations/main.go` - โปรแกรมล้างข้อมูล migration
- `migrate.bat` - ไฟล์ batch สำหรับ Windows
- เพิ่ม `make migrate` ใน Makefile

### 2. สร้างไฟล์ Migration SQL

#### ไฟล์ Migration ที่สร้าง:
1. `001_create_cargo_manifest_tables.sql` - สร้างตาราง cargo manifest
2. `002_create_draft_mawb_tables.sql` - สร้างตาราง draft MAWB
3. `003_rollback_mawb_system_integration.sql` - ไฟล์สำหรับย้อนกลับ (ไม่รันอัตโนมัติ)

### 3. ตารางที่สร้างในฐานข้อมูล

#### กลุมตาราง Cargo Manifest (2 ตาราง):

**1. ตาราง `cargo_manifest`** - ตารางหลักสำหรับข้อมูล cargo manifest
```sql
- uuid (UUID) - รหัสหลัก
- mawb_info_uuid (UUID) - เชื่อมโยงกับตาราง tbl_mawb_info
- mawb_number (VARCHAR) - หมายเลข MAWB
- port_of_discharge (VARCHAR) - ท่าขนถ่าย
- flight_no (VARCHAR) - หมายเลขเที่ยวบิน
- freight_date (DATE) - วันที่ขนส่ง
- shipper (TEXT) - ผู้ส่ง
- consignee (TEXT) - ผู้รับ
- total_ctn (VARCHAR) - จำนวนตู้รวม
- transshipment (VARCHAR) - การขนถ่าย
- status (VARCHAR) - สถานะ (Draft, Pending, Confirmed, Rejected)
- created_at, updated_at - เวลาสร้างและแก้ไข
```

**2. ตาราง `cargo_manifest_items`** - รายการสินค้าใน cargo manifest
```sql
- id (SERIAL) - รหัสหลัก
- cargo_manifest_uuid (UUID) - เชื่อมโยงกับ cargo_manifest
- hawb_no (VARCHAR) - หมายเลข HAWB
- pkgs (VARCHAR) - จำนวนหีบห่อ
- gross_weight (VARCHAR) - น้ำหนักรวม
- destination (VARCHAR) - จุดหมาย
- commodity (TEXT) - ประเภทสินค้า
- shipper_name_address (TEXT) - ชื่อที่อยู่ผู้ส่ง
- consignee_name_address (TEXT) - ชื่อที่อยู่ผู้รับ
- created_at - เวลาสร้าง
```

#### กลุ่มตาราง Draft MAWB (4 ตาราง):

**3. ตาราง `draft_mawb`** - ตารางหลักสำหรับข้อมูล draft MAWB
```sql
- uuid (UUID) - รหัสหลัก
- mawb_info_uuid (UUID) - เชื่อมโยงกับตาราง tbl_mawb_info
- customer_uuid (UUID) - รหัสลูกค้า
- airline_logo, airline_name - ข้อมูลสายการบิน
- mawb, hawb - หมายเลขเอกสาร
- shipper_name_and_address - ข้อมูลผู้ส่ง
- consignee_name_and_address - ข้อมูลผู้รับ
- issuing_carrier_agent_name_and_city - ข้อมูลตัวแทน
- accounting_information - ข้อมูลบัญชี
- airport_of_departure, airport_of_destination - สนามบินต้นทางและปลายทาง
- flight_no, flight_date - ข้อมูลเที่ยวบิน
- currency, chgs_code - ข้อมูลสกุลเงินและค่าธรรมเนียม
- declared_value_carriage, declared_value_customs - มูลค่าที่ประกาศ
- total_no_of_pieces - จำนวนชิ้นรวม
- total_gross_weight - น้ำหนักรวม
- total_chargeable_weight - น้ำหนักคิดค่าขนส่ง
- total_rate_charge, total_amount - อัตราและจำนวนเงินรวม
- status (VARCHAR) - สถานะ (Draft, Pending, Confirmed, Rejected)
- และฟิลด์อื่นๆ อีกมากมาย (รวม 40+ ฟิลด์)
```

**4. ตาราง `draft_mawb_items`** - รายการสินค้าใน draft MAWB
```sql
- id (SERIAL) - รหัสหลัก
- draft_mawb_uuid (UUID) - เชื่อมโยงกับ draft_mawb
- pieces_rcp (VARCHAR) - จำนวนชิ้น
- gross_weight (VARCHAR) - น้ำหนักรวม
- kg_lb (VARCHAR) - หน่วยน้ำหนัก
- rate_class (VARCHAR) - ประเภทอัตรา
- total_volume (DECIMAL) - ปริมาตรรวม
- chargeable_weight (DECIMAL) - น้ำหนักคิดค่าขนส่ง
- rate_charge (DECIMAL) - อัตราค่าขนส่ง
- total (DECIMAL) - จำนวนเงินรวม
- nature_and_quantity (TEXT) - ลักษณะและจำนวนสินค้า
- created_at - เวลาสร้าง
```

**5. ตาราง `draft_mawb_item_dims`** - ข้อมูลขนาดของรายการสินค้า
```sql
- id (SERIAL) - รหัสหลัก
- draft_mawb_item_id (INTEGER) - เชื่อมโยงกับ draft_mawb_items
- length (VARCHAR) - ความยาว
- width (VARCHAR) - ความกว้าง
- height (VARCHAR) - ความสูง
- count (VARCHAR) - จำนวน
- created_at - เวลาสร้าง
```

**6. ตาราง `draft_mawb_charges`** - ข้อมูลค่าธรรมเนียมต่างๆ
```sql
- id (SERIAL) - รหัสหลัก
- draft_mawb_uuid (UUID) - เชื่อมโยงกับ draft_mawb
- charge_key (VARCHAR) - ประเภทค่าธรรมเนียม
- charge_value (DECIMAL) - จำนวนเงิน
- created_at - เวลาสร้าง
```

#### ตารางระบบ Migration:

**7. ตาราง `schema_migrations`** - ติดตามการรัน migration
```sql
- version (VARCHAR) - เวอร์ชัน migration
- name (VARCHAR) - ชื่อ migration
- applied_at (TIMESTAMP) - เวลาที่รัน
```

### 4. ความสัมพันธ์ของตาราง (Foreign Keys)

#### Foreign Key Constraints ที่สร้าง (6 ความสัมพันธ์):
1. `cargo_manifest.mawb_info_uuid` → `tbl_mawb_info.uuid` (CASCADE DELETE)
2. `cargo_manifest_items.cargo_manifest_uuid` → `cargo_manifest.uuid` (CASCADE DELETE)
3. `draft_mawb.mawb_info_uuid` → `tbl_mawb_info.uuid` (CASCADE DELETE)
4. `draft_mawb_items.draft_mawb_uuid` → `draft_mawb.uuid` (CASCADE DELETE)
5. `draft_mawb_item_dims.draft_mawb_item_id` → `draft_mawb_items.id` (CASCADE DELETE)
6. `draft_mawb_charges.draft_mawb_uuid` → `draft_mawb.uuid` (CASCADE DELETE)

**หมายเหตุ CASCADE DELETE**: เมื่อลบข้อมูลในตารางหลัก ข้อมูลในตารางที่เกี่ยวข้องจะถูกลบอัตโนมัติ

### 5. Index ที่สร้างเพื่อเพิ่มประสิทธิภาพ (8 indexes)

1. `idx_cargo_manifest_mawb_info_uuid` - สำหรับ JOIN กับ tbl_mawb_info
2. `idx_cargo_manifest_status` - สำหรับกรองตามสถานะ
3. `idx_cargo_manifest_created_at` - สำหรับเรียงตามวันที่
4. `idx_cargo_manifest_items_cargo_manifest_uuid` - สำหรับ JOIN กับ cargo_manifest
5. `idx_draft_mawb_mawb_info_uuid` - สำหรับ JOIN กับ tbl_mawb_info
6. `idx_draft_mawb_status` - สำหรับกรองตามสถานะ
7. `idx_draft_mawb_created_at` - สำหรับเรียงตามวันที่
8. `idx_draft_mawb_items_draft_mawb_uuid` - สำหรับ JOIN กับ draft_mawb

## วิธีการย้อนกลับ (Rollback)

### วิธีที่ 1: ใช้ไฟล์ Rollback Migration (แนะนำ)

```bash
# รันไฟล์ rollback migration โดยตรง
psql -h [host] -U [username] -d [database] -f database/migrations/003_rollback_mawb_system_integration.sql
```

### วิธีที่ 2: ลบตารางด้วยคำสั่ง SQL

```sql
-- ลบตารางตามลำดับ (เพื่อหลีกเลี่ยง foreign key constraint error)
DROP TABLE IF EXISTS draft_mawb_charges CASCADE;
DROP TABLE IF EXISTS draft_mawb_item_dims CASCADE;
DROP TABLE IF EXISTS draft_mawb_items CASCADE;
DROP TABLE IF EXISTS draft_mawb CASCADE;
DROP TABLE IF EXISTS cargo_manifest_items CASCADE;
DROP TABLE IF EXISTS cargo_manifest CASCADE;

-- ลบข้อมูล migration records (ถ้าต้องการ)
DELETE FROM schema_migrations WHERE version IN ('001', '002');
```

### วิธีที่ 3: ใช้โปรแกรม Go ที่สร้างไว้

```bash
# ล้างข้อมูล migration records
go run cmd/clear-migrations/main.go

# จากนั้นรันไฟล์ rollback โดยตรง
```

### วิธีที่ 4: ใช้คำสั่ง SQL แบบละเอียด

```sql
-- ตรวจสอบตารางที่มีอยู่
SELECT table_name FROM information_schema.tables 
WHERE table_name IN ('cargo_manifest', 'cargo_manifest_items', 'draft_mawb', 'draft_mawb_items', 'draft_mawb_item_dims', 'draft_mawb_charges');

-- ลบ foreign key constraints ก่อน (ถ้าจำเป็น)
ALTER TABLE cargo_manifest_items DROP CONSTRAINT IF EXISTS fk_cargo_manifest_items_cargo_manifest;
ALTER TABLE cargo_manifest DROP CONSTRAINT IF EXISTS fk_cargo_manifest_mawb_info;
ALTER TABLE draft_mawb_items DROP CONSTRAINT IF EXISTS fk_draft_mawb_items_draft_mawb;
ALTER TABLE draft_mawb_item_dims DROP CONSTRAINT IF EXISTS fk_draft_mawb_item_dims_draft_mawb_item;
ALTER TABLE draft_mawb_charges DROP CONSTRAINT IF EXISTS fk_draft_mawb_charges_draft_mawb;
ALTER TABLE draft_mawb DROP CONSTRAINT IF EXISTS fk_draft_mawb_mawb_info;

-- ลบตาราง
DROP TABLE IF EXISTS draft_mawb_charges;
DROP TABLE IF EXISTS draft_mawb_item_dims;
DROP TABLE IF EXISTS draft_mawb_items;
DROP TABLE IF EXISTS draft_mawb;
DROP TABLE IF EXISTS cargo_manifest_items;
DROP TABLE IF EXISTS cargo_manifest;

-- ลบ trigger function (ถ้าไม่มีตารางอื่นใช้)
-- DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
```

## คำเตือนสำคัญ

⚠️ **การย้อนกลับจะลบข้อมูลทั้งหมดในตารางเหล่านี้อย่างถาวร**

⚠️ **ควรสำรองข้อมูลก่อนทำการย้อนกลับ**

⚠️ **ตรวจสอบให้แน่ใจว่าไม่มีระบบอื่นใช้ตารางเหล่านี้**

## การตรวจสอบหลังย้อนกลับ

```bash
# ตรวจสอบว่าตารางถูกลบแล้ว
go run cmd/verify-schema/main.go

# หรือใช้ SQL
SELECT table_name FROM information_schema.tables 
WHERE table_name IN ('cargo_manifest', 'cargo_manifest_items', 'draft_mawb', 'draft_mawb_items', 'draft_mawb_item_dims', 'draft_mawb_charges');
```

## สรุป

การทำงานนี้ได้สร้างระบบฐานข้อมูลที่สมบูรณ์สำหรับ MAWB System Integration โดยมีการจัดการ migration ที่เป็นระบบ มีความสัมพันธ์ของข้อมูลที่ถูกต้อง และมี index เพื่อเพิ่มประสิทธิภาพ พร้อมทั้งมีวิธีการย้อนกลับที่ชัดเจนและปลอดภัย