package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS billing_records (
    id INTEGER PRIMARY KEY,
    room_number TEXT NOT NULL,
    current_water REAL NOT NULL,
    previous_water REAL NOT NULL,
    water_adjustment REAL NOT NULL,
    water_usage REAL NOT NULL,
    current_electric REAL NOT NULL,
    previous_electric REAL NOT NULL,
    electric_adjustment REAL NOT NULL,
    electric_usage REAL NOT NULL,
    management_fee REAL NOT NULL,
    water_price REAL NOT NULL,
    electric_price REAL NOT NULL,
    total_water_cost REAL NOT NULL,
    total_electric_cost REAL NOT NULL,
    total_cost REAL NOT NULL,
    billing_month TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS billing_extra_fees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    billing_record_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    amount REAL NOT NULL,
    FOREIGN KEY (billing_record_id) REFERENCES billing_records(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS total_meter_records (
    id INTEGER PRIMARY KEY,
    month TEXT NOT NULL UNIQUE,
    water_reading REAL NOT NULL,
    electric_reading REAL NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_billing_records_month ON billing_records(billing_month);
CREATE INDEX IF NOT EXISTS idx_billing_records_room_created ON billing_records(room_number, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_billing_extra_fees_record ON billing_extra_fees(billing_record_id);
`

// OpenSQLite 初始化 SQLite 连接和不可变的表结构。单连接可避免嵌入式 SQLite
// 在高频写入时出现锁竞争，busy_timeout 则为临时锁占用留下短暂等待窗口。
func OpenSQLite(databaseFile string) (*sql.DB, error) {
	if databaseFile == "" {
		return nil, fmt.Errorf("sqlite 数据库路径不能为空")
	}
	if err := os.MkdirAll(filepath.Dir(databaseFile), 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sql.Open("sqlite", databaseFile)
	if err != nil {
		return nil, fmt.Errorf("打开 sqlite 数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000; PRAGMA journal_mode = WAL;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("初始化 sqlite 参数失败: %w", err)
	}
	if _, err := db.Exec(sqliteSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("创建 sqlite 表结构失败: %w", err)
	}
	if err := ensureBillingRecordUniqueness(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ensureBillingRecordUniqueness(db *sql.DB) error {
	var roomNumber, month string
	var count int
	err := db.QueryRow(`SELECT room_number, billing_month, COUNT(*) FROM billing_records GROUP BY room_number, billing_month HAVING COUNT(*) > 1 LIMIT 1`).Scan(&roomNumber, &month, &count)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查重复账单失败: %w", err)
	}
	if err == nil {
		return fmt.Errorf("发现重复账单：房号 %q 在 %s 有 %d 条记录；请先合并后再启动服务", roomNumber, month, count)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_billing_records_room_month ON billing_records(room_number, billing_month)`); err != nil {
		return fmt.Errorf("创建账单唯一索引失败: %w", err)
	}
	return nil
}
