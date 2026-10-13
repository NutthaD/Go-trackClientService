# Go-trackClientService

แอปให้ tester ค้นหาและติดตามสถานะผู้ว่าจ้าง (Client) กับงานบริการ (Service)

- Backend: Go
- Database: MySQL Server (ติดตั้งบนเครื่อง ไม่ใช้ Docker)
- Frontend: React (ยังไม่สร้างในขั้นต้น)

## ขั้นที่ทำแล้ว

1. โครงโฟลเดอร์ + Go module + สคริปต์สร้างฐานข้อมูลบน MySQL Server
2. API ตรวจสถานะ (`/health`, `/health/db`)
3. API จัดการผู้ว่าจ้างและ service (`/clients`)

## รัน backend

จากโฟลเดอร์ `backend`:

```bash
go run ./cmd/api
```

แล้วเปิด:

- http://127.0.0.1:8080/health — โปรเซส Go ทำงาน
- http://127.0.0.1:8080/health/db — ต่อ MySQL ได้หรือไม่

## ตารางและ API ผู้ว่าจ้าง

หลังสร้างฐานข้อมูล `trackclient` แล้ว ให้รัน `backend/migrations/001_create_clients_services.sql` ใน DBeaver เพื่อสร้างตาราง `clients` และ `services` ตาราง service เชื่อมกับ client ด้วย foreign key และเก็บรหัสจาก Excel เช่น `123_00` ไว้ใน `client_service_id`

API ทุกเส้นทางคืนข้อมูล JSON ยกเว้นการลบที่สำเร็จจะคืน HTTP `204` โดยไม่มี body:

| Method | Path | หน้าที่ |
|---|---|---|
| `POST` | `/clients` | เพิ่ม client พร้อม services ได้ในคำขอเดียว |
| `GET` | `/clients?limit=50&offset=0` | อ่าน client พร้อม services (limit สูงสุด 100) |
| `GET` | `/clients/{client_id}` | อ่าน client หนึ่งรายพร้อม services |
| `PATCH` | `/clients/{client_id}` | เปลี่ยนชื่อ client เท่านั้น |
| `DELETE` | `/clients/{client_id}` | ลบ client ได้เมื่อไม่มี service เหลือ |
| `POST` | `/clients/{client_id}/services` | เพิ่ม service ให้ client |
| `PATCH` | `/clients/{client_id}/services/{client_service_id}` | เปลี่ยนชื่อ service เท่านั้น |
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
      "serv_name": "ระบบ A"
    }
  ]
}
```

`client_service_id` จะถูกสร้างจาก `client_id` และ `service_id` เป็น `123_00` ถ้าส่งมาด้วย API จะตรวจว่าตรงกันหรือไม่ การเพิ่ม client และ services ใช้ transaction: ถ้าส่วนใดส่วนหนึ่งผิดพลาด จะยกเลิกการเพิ่มทั้งชุด

การแก้ไข client และ service แยก endpoint กัน เพื่อให้เปลี่ยนเฉพาะข้อมูลที่ต้องการได้ การลบ client ที่ยังมี services จะตอบ `409 Conflict`; ต้องลบ services ก่อน ข้อมูลซ้ำตอบ `409 Conflict` และ ID ที่ไม่พบตอบ `404 Not Found`

> Migration ใช้ `CREATE TABLE IF NOT EXISTS` เพื่อไม่สร้างตารางซ้ำ แต่ไม่ได้แก้ตารางเดิม หากสร้างตารางด้วยมือไว้แล้ว ให้ตรวจว่าคอลัมน์และชนิดตรงกับ migration ก่อนใช้งาน API

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
