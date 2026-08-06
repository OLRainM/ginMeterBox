package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ginMeterBox/models"
)

// MigrateJSONToSQLiteIfNeeded 只在全新 SQLite 数据库中执行一次性迁移。
// JSON 始终保留为可人工核验和回滚的源备份，不会在应用运行期间被覆盖。
func MigrateJSONToSQLiteIfNeeded(db *sql.DB, billingFile, totalMeterFile string) error {
	var billingCount, totalMeterCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM billing_records`).Scan(&billingCount); err != nil {
		return err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM total_meter_records`).Scan(&totalMeterCount); err != nil {
		return err
	}
	if billingCount > 0 || totalMeterCount > 0 {
		return nil
	}

	billingRecords, err := readBillingRecordsJSON(billingFile)
	if err != nil {
		return err
	}
	totalMeterRecords, err := readTotalMeterRecordsJSON(totalMeterFile)
	if err != nil {
		return err
	}
	if len(billingRecords) == 0 && len(totalMeterRecords) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i := range billingRecords {
		id, err := insertBillingRecord(tx, &billingRecords[i], true)
		if err != nil {
			return fmt.Errorf("迁移账单 ID=%d 失败: %w", billingRecords[i].ID, err)
		}
		if err := replaceExtraFees(tx, id, billingRecords[i].ExtraFees); err != nil {
			return fmt.Errorf("迁移账单 ID=%d 额外费用失败: %w", billingRecords[i].ID, err)
		}
	}
	for _, record := range totalMeterRecords {
		if _, err := tx.Exec(`INSERT INTO total_meter_records (id, month, water_reading, electric_reading, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, record.ID, record.Month, record.WaterReading, record.ElectricReading, record.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"), record.UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00")); err != nil {
			return fmt.Errorf("迁移总表月份=%s 失败: %w", record.Month, err)
		}
	}
	return tx.Commit()
}

// BackfillLegacyMasterBillsToTotalMeters 将旧账单中 roomNumber="总表" 的当月读数
// 补写到独立总表表。仅写入缺少的月份，绝不覆盖用户在总表页面手工维护的数据。
// 这使历史数据可以被新的 /total-meter/month API 使用，同时仍保留原始账单审计记录。
func BackfillLegacyMasterBillsToTotalMeters(db *sql.DB) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
        SELECT b.billing_month, b.current_water, b.current_electric, b.created_at, b.updated_at
        FROM billing_records b
        LEFT JOIN total_meter_records t ON t.month = b.billing_month
        WHERE b.room_number = '总表' AND t.month IS NULL
        ORDER BY b.billing_month
    `)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type legacyRecord struct {
		month, createdAt, updatedAt string
		water, electric             float64
	}
	legacyRecords := make([]legacyRecord, 0)
	for rows.Next() {
		var record legacyRecord
		if err := rows.Scan(&record.month, &record.water, &record.electric, &record.createdAt, &record.updatedAt); err != nil {
			return 0, err
		}
		legacyRecords = append(legacyRecords, record)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}

	for _, record := range legacyRecords {
		if _, err := tx.Exec(`INSERT INTO total_meter_records (month, water_reading, electric_reading, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, record.month, record.water, record.electric, record.createdAt, record.updatedAt); err != nil {
			return 0, fmt.Errorf("回填历史总表月份=%s 失败: %w", record.month, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(legacyRecords), nil
}

func readBillingRecordsJSON(filename string) ([]models.BillingRecord, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.BillingRecord{}, nil
		}
		return nil, fmt.Errorf("读取账单 JSON 失败: %w", err)
	}
	var records []models.BillingRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("解析账单 JSON 失败: %w", err)
	}
	return records, nil
}

func readTotalMeterRecordsJSON(filename string) ([]models.TotalMeterRecord, error) {
	if filename == "" {
		return []models.TotalMeterRecord{}, nil
	}
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		if os.IsNotExist(err) {
			return []models.TotalMeterRecord{}, nil
		}
		return nil, fmt.Errorf("读取总表 JSON 失败: %w", err)
	}
	var records []models.TotalMeterRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("解析总表 JSON 失败: %w", err)
	}
	return records, nil
}
