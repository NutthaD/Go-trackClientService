# Track Client Service frontend

หน้า React สำหรับดู ค้นหา กรอง และแก้ไขข้อมูล UAT ของ service

## เริ่มใช้งาน

1. เปิด Go API ใน terminal หนึ่ง จากโฟลเดอร์ `backend`:

   ```bash
   go run ./cmd/api
   ```

2. เปิด frontend ใน terminal อีกหน้าต่าง:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

3. เปิด `http://localhost:5173`

Vite จะ proxy คำขอ `/api` ไปยัง Go API ที่ `http://127.0.0.1:8080` โดยอัตโนมัติ
