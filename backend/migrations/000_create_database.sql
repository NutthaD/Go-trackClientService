-- รันครั้งแรกบน MySQL Server ของเครื่องนั้น
-- mysql -u root -p < backend/migrations/000_create_database.sql

CREATE DATABASE IF NOT EXISTS trackclient
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
