package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Open สร้าง connection pool — ยังไม่คุยกับ MySQL จนกว่าจะ Ping
func Open(dsn string) (*sql.DB, error) {
	return sql.Open("mysql", dsn)
}
