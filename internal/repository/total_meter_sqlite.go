package repository

import (
	"database/sql"
	"fmt"
	"time"

	"ginMeterBox/internal/models"
)

// TotalMeterSQLiteRepo 是 TotalMeterRepo 的 SQLite 实现。独立总表数据不与历史“总表”账单混写，
// 从而不会在住户补差计算中被重复统计。
type TotalMeterSQLiteRepo struct {
	db *sql.DB
}

func NewTotalMeterSQLiteRepo(db *sql.DB) *TotalMeterSQLiteRepo {
	return &TotalMeterSQLiteRepo{db: db}
}

func scanTotalMeterRecord(row interface{ Scan(...any) error }) (models.TotalMeterRecord, error) {
	var record models.TotalMeterRecord
	var createdAt, updatedAt string
	if err := row.Scan(&record.ID, &record.Month, &record.WaterReading, &record.ElectricReading, &createdAt, &updatedAt); err != nil {
		return models.TotalMeterRecord{}, err
	}
	var err error
	record.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return models.TotalMeterRecord{}, fmt.Errorf("解析总表创建时间失败: %w", err)
	}
	record.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return models.TotalMeterRecord{}, fmt.Errorf("解析总表更新时间失败: %w", err)
	}
	return record, nil
}

func (r *TotalMeterSQLiteRepo) GetAll() []models.TotalMeterRecord {
	rows, err := r.db.Query(`SELECT id, month, water_reading, electric_reading, created_at, updated_at FROM total_meter_records ORDER BY month DESC`)
	if err != nil {
		return []models.TotalMeterRecord{}
	}
	defer rows.Close()
	records := make([]models.TotalMeterRecord, 0)
	for rows.Next() {
		record, err := scanTotalMeterRecord(rows)
		if err != nil {
			return []models.TotalMeterRecord{}
		}
		records = append(records, record)
	}
	return records
}

func (r *TotalMeterSQLiteRepo) GetByMonth(month string) (*models.TotalMeterRecord, error) {
	record, err := scanTotalMeterRecord(r.db.QueryRow(`SELECT id, month, water_reading, electric_reading, created_at, updated_at FROM total_meter_records WHERE month = ?`, month))
	if err == sql.ErrNoRows {
		return nil, ErrTotalMeterRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *TotalMeterSQLiteRepo) Create(record *models.TotalMeterRecord) error {
	now := time.Now()
	record.CreatedAt, record.UpdatedAt = now, now
	result, err := r.db.Exec(`INSERT INTO total_meter_records (month, water_reading, electric_reading, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, record.Month, record.WaterReading, record.ElectricReading, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		// 只有同一月份的唯一约束冲突才映射为可预期的业务错误，其他数据库故障必须上抛。
		if existing, lookupErr := r.GetByMonth(record.Month); lookupErr == nil && existing != nil {
			return ErrTotalMeterMonthExists
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	record.ID = int(id)
	return nil
}

func (r *TotalMeterSQLiteRepo) Update(month string, record *models.TotalMeterRecord) error {
	existing, err := r.GetByMonth(month)
	if err != nil {
		return err
	}
	record.ID, record.Month, record.CreatedAt, record.UpdatedAt = existing.ID, month, existing.CreatedAt, time.Now()
	result, err := r.db.Exec(`UPDATE total_meter_records SET water_reading = ?, electric_reading = ?, updated_at = ? WHERE month = ?`, record.WaterReading, record.ElectricReading, record.UpdatedAt.Format(time.RFC3339Nano), month)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return ErrTotalMeterRecordNotFound
	}
	return nil
}

func (r *TotalMeterSQLiteRepo) Delete(month string) error {
	result, err := r.db.Exec(`DELETE FROM total_meter_records WHERE month = ?`, month)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return ErrTotalMeterRecordNotFound
	}
	return nil
}
