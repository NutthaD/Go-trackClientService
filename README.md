# Go-trackClientService

แอปให้ tester ค้นหาและติดตามสถานะผู้ว่าจ้าง (Client) กับงานบริการ (Service)

- Backend: Go
- Database: MySQL Server (ติดตั้งบนเครื่อง ไม่ใช้ Docker)
- Frontend: React (ยังไม่สร้างในขั้นต้น)

## ขั้นที่ทำแล้ว

1. โครงโฟลเดอร์ + Go module + สคริปต์สร้างฐานข้อมูลบน MySQL Server
2. API เล็กๆ (`/health`, `/health/db`) ต่อ MySQL ด้วย Ping

## รัน backend

จากโฟลเดอร์ `backend`:

```bash
go run ./cmd/api
```

แล้วเปิด:

- http://127.0.0.1:8080/health — โปรเซส Go ทำงาน
- http://127.0.0.1:8080/health/db — ต่อ MySQL ได้หรือไม่

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

Go จะอ่านค่าเหล่านี้ตอนต่อฐานข้อมูล (ยังไม่ได้เขียนโค้ดต่อในขั้นนี้)
