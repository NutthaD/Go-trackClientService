# Go-trackClientService

แอปให้ tester ค้นหาและติดตามสถานะผู้ว่าจ้าง (Client) กับงานบริการ (Service)

- Backend: Go
- Database: MySQL Server (ติดตั้งบนเครื่อง ไม่ใช้ Docker)
- Frontend: React + TypeScript + Vite

## ขั้นที่ทำแล้ว

1. โครงโฟลเดอร์ + Go module + สคริปต์สร้างฐานข้อมูลบน MySQL Server
2. API ตรวจสถานะ (`/health`, `/health/db`)
3. API จัดการผู้ว่าจ้างและ service (`/clients`)
4. Frontend เบื้องต้นสำหรับค้นหา ติดตาม และแก้ไขข้อมูล service

## รัน frontend

ต้องเปิด Go API ไว้ก่อน โดยรัน backend ตามขั้นตอนด้านบน จากนั้นเปิดอีกหน้าต่าง Terminal:

```bash
cd frontend
npm install
npm run dev
```

เปิด URL ที่ Vite แสดง (ปกติคือ http://localhost:5173) หน้าเว็บจะเรียก Go API ผ่าน proxy ที่ตั้งไว้ใน Vite โดยส่ง `/api/clients` ไปยัง `http://127.0.0.1:8080/clients` จึงไม่ต้องตั้งค่า CORS สำหรับการพัฒนาในเครื่องนี้

หน้าแรกแสดงสรุปสถานะ UAT, ค้นหาด้วยชื่อหรือรหัส, กรองตามสถานะ และแก้ไขข้อมูลบริการได้จากปุ่ม “แก้ไข” การแก้ไขจะบันทึกผ่าน `PATCH /clients/{client_id}/services/{client_service_id}` ทันที

> การตั้งค่า proxy นี้ใช้สำหรับรัน frontend และ backend บนเครื่องเดียวกันเพื่อพัฒนาเท่านั้น การเปิดให้เครื่องอื่นใช้งานหรือ deploy แยกเครื่องจะต้องกำหนด URL และวิธีเชื่อมต่อ frontend/API เพิ่มเติม

## รัน backend

จากโฟลเดอร์ `backend`:

```bash
go run ./cmd/api
```

แล้วเปิด:

- http://127.0.0.1:8080/health — โปรเซส Go ทำงาน
- http://127.0.0.1:8080/health/db — ต่อ MySQL ได้หรือไม่

## ตารางและ API ผู้ว่าจ้าง

หลังสร้างฐานข้อมูล `trackclient` แล้ว ให้รัน `backend/migrations/001_create_clients_services.sql` ใน DBeaver เพื่อสร้างตาราง `clients` และ `services` ตาราง service เชื่อมกับ client ด้วย foreign key และเก็บรหัสจาก Excel เช่น `123_00` ไว้ใน `client_service_id` พร้อมฟิลด์ UAT และ Production

ถ้าฐานข้อมูลเดิมสร้างจาก migration รุ่นก่อนที่ยังไม่มีฟิลด์ UAT ให้รัน `backend/migrations/002_add_service_uat_fields.sql` เพียงครั้งเดียวแทนการรัน migration 001 ซ้ำ Migration 002 ใช้สำหรับ schema เก่าเท่านั้น; ฐานข้อมูลที่อัปเดตฟิลด์ด้วยตนเองแล้วไม่ต้องรันซ้ำ

API ทุกเส้นทางคืนข้อมูล JSON ยกเว้นการลบที่สำเร็จจะคืน HTTP `204` โดยไม่มี body:

| Method | Path | หน้าที่ |
|---|---|---|
| `POST` | `/clients` | เพิ่ม client พร้อม services ได้ในคำขอเดียว |
| `GET` | `/clients?limit=50&offset=0` | อ่าน client พร้อม services (limit สูงสุด 100) |
| `GET` | `/clients/{client_id}` | อ่าน client หนึ่งรายพร้อม services |
| `PATCH` | `/clients/{client_id}` | เปลี่ยนชื่อ client เท่านั้น |
| `DELETE` | `/clients/{client_id}` | ลบ client ได้เมื่อไม่มี service เหลือ |
| `POST` | `/clients/{client_id}/services` | เพิ่ม service ให้ client |
| `PATCH` | `/clients/{client_id}/services/{client_service_id}` | เปลี่ยนเฉพาะฟิลด์ service ที่ส่งมา |
| `DELETE` | `/clients/{client_id}/services/{client_service_id}` | ลบ service หนึ่งรายการ |

ตัวอย่างเพิ่ม client พร้อม service:

```json
{
  "client_id": "123",
  "client_short_name": "บริษัทตัวอย่าง",
  "services": [
    {
      "client_service_id": "123_00",
      "service_id": "00",
      "serv_name": "ระบบ A",
      "uat_status": "Waiting test",
      "tested_by": "Tester A",
      "pos_os_nss": "Windows",
      "testcase_result_link": "https://example.com/test-result",
      "week": "1",
      "uat_submitted_date": "2026-10-24",
      "service_type": "NPE",
      "production_status": "NOT_PROD"
    }
  ]
}
```

`client_service_id` จะถูกสร้างจาก `client_id` และ `service_id` เป็น `123_00` ถ้าส่งมาด้วย API จะตรวจว่าตรงกันหรือไม่ การเพิ่ม client และ services ใช้ transaction: ถ้าส่วนใดส่วนหนึ่งผิดพลาด จะยกเลิกการเพิ่มทั้งชุด

ฟิลด์ `uat_status` รับค่า `Waiting test`, `Block`, `Block by defect`, `Defect`, `Pass`, `Cancel` หรือ `Hold testing` และถ้าไม่ส่งจะเริ่มเป็น `Waiting test` ส่วน `production_status` รับ `PROD`, `NOT_PROD` หรือ `NOT_APPLICABLE` และค่าเริ่มต้นคือ `NOT_PROD` ฟิลด์ข้อมูลเพิ่มเติมที่ไม่ส่งสามารถว่างได้ วันที่ `uat_submitted_date` ใช้รูปแบบ `YYYY-MM-DD`

ตัวอย่างแก้เฉพาะ UAT status โดยไม่เปลี่ยนฟิลด์อื่น:

```http
PATCH /clients/123/services/123_00
Content-Type: application/json
```

```json
{
  "uat_status": "Pass"
}
```

`PATCH` รับฟิลด์ service ใดก็ได้ที่ต้องการเปลี่ยน เช่น `serv_name`, `uat_status`, `tested_by`, `pos_os_nss`, `testcase_result_link`, `week`, `uat_submitted_date`, `service_type` และ `production_status` ฟิลด์ที่ไม่ส่งจะคงค่าเดิม ส่ง `""` ในฟิลด์ข้อความที่อนุญาตให้ว่างเพื่อล้างค่านั้นเป็น `NULL`

การแก้ไข client และ service แยก endpoint กัน เพื่อให้เปลี่ยนเฉพาะข้อมูลที่ต้องการได้ การลบ client ที่ยังมี services จะตอบ `409 Conflict`; ต้องลบ services ก่อน ข้อมูลซ้ำตอบ `409 Conflict` และ ID ที่ไม่พบตอบ `404 Not Found`

> Migration 001 ใช้ `CREATE TABLE IF NOT EXISTS` เพื่อไม่สร้างตารางซ้ำ แต่ไม่ได้แก้ตารางเดิม หากตารางถูกสร้างด้วยมือหรือจาก schema รุ่นเก่า ให้ตรวจว่ามีคอลัมน์ตามรายการข้างต้นก่อนใช้งาน API; ใช้ migration 002 เฉพาะกับ schema รุ่นเก่าที่ยังไม่มีคอลัมน์ UAT

## MySQL บนแต่ละเครื่อง

1. ติดตั้ง [MySQL Server](https://dev.mysql.com/downloads/mysql/) (หรือ MySQL จาก XAMPP / MySQL Installer)
2. สร้างฐานข้อมูล:

```bash
mysql -u root -p < backend/migrations/000_create_database.sql
```

หรือใน MySQL Workbench / บรรทัดคำสั่ง:

```sql
CREATE DATABASE IF NOT EXISTS trackclient
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

3. คัดลอก `.env.example` เป็น `.env` แล้วใส่ host / user / password ของเครื่องนั้น

Go จะอ่านค่าเหล่านี้ตอนเริ่มทำงาน ค่า `.env` ไม่ควรถูก commit ขึ้น Git
ค่า `HTTP_HOST=127.0.0.1` เปิดให้เข้า API ได้เฉพาะเครื่องที่รันอยู่ หากต้องการให้ tester ในเครือข่ายส่วนตัวเข้าถึง ให้ตั้ง `HTTP_HOST=0.0.0.0` บนเครื่องที่รัน API และตั้ง Firewall ให้รับเฉพาะเครือข่ายที่ไว้ใจได้ ห้ามเปิด MySQL port ออกสู่อินเทอร์เน็ตโดยตรง
