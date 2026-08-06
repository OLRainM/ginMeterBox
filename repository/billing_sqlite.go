package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"ginMeterBox/models"
)

// BillingSQLiteRepo 是 BillingRepo 的 SQLite 实现。账单及其额外费用始终在同一事务中保存。
type BillingSQLiteRepo struct {
	db *sql.DB
}

func NewBillingSQLiteRepo(db *sql.DB) *BillingSQLiteRepo {
	return &BillingSQLiteRepo{db: db}
}

const billingColumns = `id, room_number, current_water, previous_water, water_adjustment, water_usage,
current_electric, previous_electric, electric_adjustment, electric_usage, management_fee,
water_price, electric_price, total_water_cost, total_electric_cost, total_cost,
billing_month, created_at, updated_at`

func scanBillingRecord(row interface{ Scan(...any) error }) (models.BillingRecord, error) {
	var record models.BillingRecord
	var createdAt, updatedAt string
	err := row.Scan(
		&record.ID, &record.RoomNumber, &record.CurrentWater, &record.PreviousWater,
		&record.WaterAdjustment, &record.WaterUsage, &record.CurrentElectric, &record.PreviousElectric,
		&record.ElectricAdjustment, &record.ElectricUsage, &record.ManagementFee, &record.WaterPrice,
		&record.ElectricPrice, &record.TotalWaterCost, &record.TotalElectricCost, &record.TotalCost,
		&record.BillingMonth, &createdAt, &updatedAt,
	)
	if err != nil {
		return models.BillingRecord{}, err
	}
	var parseErr error
	record.CreatedAt, parseErr = time.Parse(time.RFC3339Nano, createdAt)
	if parseErr != nil {
		return models.BillingRecord{}, fmt.Errorf("解析账单创建时间失败: %w", parseErr)
	}
	record.UpdatedAt, parseErr = time.Parse(time.RFC3339Nano, updatedAt)
	if parseErr != nil {
		return models.BillingRecord{}, fmt.Errorf("解析账单更新时间失败: %w", parseErr)
	}
	return record, nil
}

func (r *BillingSQLiteRepo) loadExtraFees(record *models.BillingRecord) error {
	rows, err := r.db.Query(`SELECT name, amount FROM billing_extra_fees WHERE billing_record_id = ? ORDER BY id`, record.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	record.ExtraFees = []models.ExtraFee{}
	for rows.Next() {
		var fee models.ExtraFee
		if err := rows.Scan(&fee.Name, &fee.Amount); err != nil {
			return err
		}
		record.ExtraFees = append(record.ExtraFees, fee)
	}
	return rows.Err()
}

func (r *BillingSQLiteRepo) loadRecords(rows *sql.Rows) []models.BillingRecord {
	records := make([]models.BillingRecord, 0)
	for rows.Next() {
		record, err := scanBillingRecord(rows)
		if err != nil {
			rows.Close()
			return []models.BillingRecord{}
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return []models.BillingRecord{}
	}
	// 必须先释放 rows 占用的连接，再查询额外费用；SQLite 被刻意限制为单连接以避免写锁竞争。
	if err := rows.Close(); err != nil {
		return []models.BillingRecord{}
	}
	for i := range records {
		if err := r.loadExtraFees(&records[i]); err != nil {
			return []models.BillingRecord{}
		}
	}
	return records
}

func (r *BillingSQLiteRepo) GetAll() []models.BillingRecord {
	rows, err := r.db.Query(`SELECT ` + billingColumns + ` FROM billing_records ORDER BY billing_month DESC, id DESC`)
	if err != nil {
		return []models.BillingRecord{}
	}
	return r.loadRecords(rows)
}

func (r *BillingSQLiteRepo) GetByID(id int) (*models.BillingRecord, error) {
	record, err := scanBillingRecord(r.db.QueryRow(`SELECT `+billingColumns+` FROM billing_records WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadExtraFees(&record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *BillingSQLiteRepo) GetByMonth(month string) []models.BillingRecord {
	rows, err := r.db.Query(`SELECT `+billingColumns+` FROM billing_records WHERE billing_month = ? ORDER BY id`, month)
	if err != nil {
		return []models.BillingRecord{}
	}
	return r.loadRecords(rows)
}

func (r *BillingSQLiteRepo) GetByIDs(ids []int) []models.BillingRecord {
	if len(ids) == 0 {
		return []models.BillingRecord{}
	}
	args := make([]any, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i], placeholders[i] = id, "?"
	}
	rows, err := r.db.Query(`SELECT `+billingColumns+` FROM billing_records WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return []models.BillingRecord{}
	}
	records := r.loadRecords(rows)
	position := make(map[int]int, len(ids))
	for i, id := range ids {
		position[id] = i
	}
	sort.SliceStable(records, func(i, j int) bool { return position[records[i].ID] < position[records[j].ID] })
	return records
}

func (r *BillingSQLiteRepo) GetLatestByRoomNumber(roomNumber string) (*models.BillingRecord, error) {
	record, err := scanBillingRecord(r.db.QueryRow(`SELECT `+billingColumns+` FROM billing_records WHERE room_number = ? ORDER BY billing_month DESC, created_at DESC, id DESC LIMIT 1`, roomNumber))
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadExtraFees(&record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *BillingSQLiteRepo) GetByRoomAndMonth(roomNumber, month string) (*models.BillingRecord, error) {
	record, err := scanBillingRecord(r.db.QueryRow(`SELECT `+billingColumns+` FROM billing_records WHERE room_number = ? AND billing_month = ?`, roomNumber, month))
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadExtraFees(&record); err != nil {
		return nil, err
	}
	return &record, nil
}

func billingPeriodExists(tx *sql.Tx, roomNumber, month string, excludeID int) (bool, error) {
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM billing_records WHERE room_number = ? AND billing_month = ? AND id != ?`, roomNumber, month, excludeID).Scan(&count); err != nil {
		return false, err
	}
	return count != 0, nil
}

func insertBillingRecord(tx *sql.Tx, record *models.BillingRecord, preserveID bool) (int, error) {
	columns := `room_number, current_water, previous_water, water_adjustment, water_usage,
        current_electric, previous_electric, electric_adjustment, electric_usage, management_fee,
        water_price, electric_price, total_water_cost, total_electric_cost, total_cost,
        billing_month, created_at, updated_at`
	placeholders := `?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?`
	args := []any{
		record.RoomNumber, record.CurrentWater, record.PreviousWater, record.WaterAdjustment, record.WaterUsage,
		record.CurrentElectric, record.PreviousElectric, record.ElectricAdjustment, record.ElectricUsage,
		record.ManagementFee, record.WaterPrice, record.ElectricPrice, record.TotalWaterCost,
		record.TotalElectricCost, record.TotalCost, record.BillingMonth,
		record.CreatedAt.Format(time.RFC3339Nano), record.UpdatedAt.Format(time.RFC3339Nano),
	}
	if preserveID {
		columns = "id, " + columns
		placeholders = "?, " + placeholders
		args = append([]any{record.ID}, args...)
	}
	result, err := tx.Exec(`INSERT INTO billing_records (`+columns+`) VALUES (`+placeholders+`)`, args...)
	if err != nil {
		return 0, err
	}
	if preserveID {
		return record.ID, nil
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func replaceExtraFees(tx *sql.Tx, recordID int, fees []models.ExtraFee) error {
	if _, err := tx.Exec(`DELETE FROM billing_extra_fees WHERE billing_record_id = ?`, recordID); err != nil {
		return err
	}
	for _, fee := range fees {
		if _, err := tx.Exec(`INSERT INTO billing_extra_fees (billing_record_id, name, amount) VALUES (?, ?, ?)`, recordID, fee.Name, fee.Amount); err != nil {
			return err
		}
	}
	return nil
}

func (r *BillingSQLiteRepo) Create(record *models.BillingRecord) error {
	now := time.Now()
	record.CreatedAt, record.UpdatedAt = now, now
	record.CalculateCosts()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	exists, err := billingPeriodExists(tx, record.RoomNumber, record.BillingMonth, 0)
	if err != nil {
		return err
	}
	if exists {
		return ErrBillingPeriodExists
	}
	id, err := insertBillingRecord(tx, record, false)
	if err != nil {
		return err
	}
	if err := replaceExtraFees(tx, id, record.ExtraFees); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	record.ID = id
	return nil
}

func (r *BillingSQLiteRepo) BatchCreate(records []models.BillingRecord) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now()
	for i := range records {
		records[i].CreatedAt, records[i].UpdatedAt = now, now
		records[i].CalculateCosts()
		exists, err := billingPeriodExists(tx, records[i].RoomNumber, records[i].BillingMonth, 0)
		if err != nil {
			return err
		}
		if exists {
			return ErrBillingPeriodExists
		}
		id, err := insertBillingRecord(tx, &records[i], false)
		if err != nil {
			return err
		}
		if err := replaceExtraFees(tx, id, records[i].ExtraFees); err != nil {
			return err
		}
		records[i].ID = id
	}
	return tx.Commit()
}

func (r *BillingSQLiteRepo) Update(id int, record *models.BillingRecord) error {
	existing, err := r.GetByID(id)
	if err != nil {
		return err
	}
	record.ID, record.CreatedAt, record.UpdatedAt = id, existing.CreatedAt, time.Now()
	record.CalculateCosts()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	exists, err := billingPeriodExists(tx, record.RoomNumber, record.BillingMonth, id)
	if err != nil {
		return err
	}
	if exists {
		return ErrBillingPeriodExists
	}
	result, err := tx.Exec(`UPDATE billing_records SET room_number=?, current_water=?, previous_water=?, water_adjustment=?, water_usage=?, current_electric=?, previous_electric=?, electric_adjustment=?, electric_usage=?, management_fee=?, water_price=?, electric_price=?, total_water_cost=?, total_electric_cost=?, total_cost=?, billing_month=?, updated_at=? WHERE id=?`,
		record.RoomNumber, record.CurrentWater, record.PreviousWater, record.WaterAdjustment, record.WaterUsage,
		record.CurrentElectric, record.PreviousElectric, record.ElectricAdjustment, record.ElectricUsage,
		record.ManagementFee, record.WaterPrice, record.ElectricPrice, record.TotalWaterCost, record.TotalElectricCost,
		record.TotalCost, record.BillingMonth, record.UpdatedAt.Format(time.RFC3339Nano), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return ErrRecordNotFound
	}
	if err := replaceExtraFees(tx, id, record.ExtraFees); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *BillingSQLiteRepo) Delete(id int) error {
	count, err := r.BatchDelete([]int{id})
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func distinctIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func (r *BillingSQLiteRepo) BatchDelete(ids []int) (int, error) {
	ids = distinctIDs(ids)
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i := range ids {
		args[i] = ids[i]
	}
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`DELETE FROM billing_records WHERE id IN (`+placeholders+`)`, args...)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if int(count) != len(ids) {
		return 0, ErrRecordNotFound
	}
	return int(count), tx.Commit()
}

func (r *BillingSQLiteRepo) BatchUpdateAdjustments(ids []int, waterAdjustment, electricAdjustment *float64) (int, error) {
	ids = distinctIDs(ids)
	if len(ids) == 0 {
		return 0, nil
	}
	records := r.GetByIDs(ids)
	if len(records) != len(ids) {
		return 0, ErrRecordNotFound
	}
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	for _, record := range records {
		if waterAdjustment != nil {
			record.WaterAdjustment = *waterAdjustment
		}
		if electricAdjustment != nil {
			record.ElectricAdjustment = *electricAdjustment
		}
		if !hasNonNegativeUsage(record) {
			return 0, ErrInvalidUsage
		}
		record.UpdatedAt = time.Now()
		record.CalculateCosts()
		if _, err := tx.Exec(`UPDATE billing_records SET water_adjustment=?, electric_adjustment=?, water_usage=?, electric_usage=?, total_water_cost=?, total_electric_cost=?, total_cost=?, updated_at=? WHERE id=?`, record.WaterAdjustment, record.ElectricAdjustment, record.WaterUsage, record.ElectricUsage, record.TotalWaterCost, record.TotalElectricCost, record.TotalCost, record.UpdatedAt.Format(time.RFC3339Nano), record.ID); err != nil {
			return 0, err
		}
	}
	return len(records), tx.Commit()
}

func (r *BillingSQLiteRepo) BatchUpdateWaterReadings(updates []WaterReadingUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	ids := make([]int, len(updates))
	readingByID := make(map[int]float64, len(updates))
	for i, update := range updates {
		if _, exists := readingByID[update.ID]; exists {
			return ErrRecordNotFound
		}
		ids[i] = update.ID
		readingByID[update.ID] = update.CurrentWater
	}
	records := r.GetByIDs(ids)
	if len(records) != len(ids) {
		return ErrRecordNotFound
	}
	for i := range records {
		records[i].CurrentWater = readingByID[records[i].ID]
		if !hasNonNegativeUsage(records[i]) {
			return ErrInvalidUsage
		}
		records[i].UpdatedAt = time.Now()
		records[i].CalculateCosts()
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, record := range records {
		if _, err := tx.Exec(`UPDATE billing_records SET current_water=?, water_usage=?, total_water_cost=?, total_cost=?, updated_at=? WHERE id=?`, record.CurrentWater, record.WaterUsage, record.TotalWaterCost, record.TotalCost, record.UpdatedAt.Format(time.RFC3339Nano), record.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *BillingSQLiteRepo) BatchSetExtraFees(ids []int, extraFees []models.ExtraFee, mode string) (int, error) {
	ids = distinctIDs(ids)
	if len(ids) == 0 {
		return 0, nil
	}
	records := r.GetByIDs(ids)
	if len(records) != len(ids) {
		return 0, ErrRecordNotFound
	}
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	for _, record := range records {
		if mode == "replace" {
			record.ExtraFees = append([]models.ExtraFee(nil), extraFees...)
		} else {
			record.ExtraFees = append(record.ExtraFees, extraFees...)
		}
		record.UpdatedAt = time.Now()
		record.CalculateCosts()
		if _, err := tx.Exec(`UPDATE billing_records SET total_cost=?, updated_at=? WHERE id=?`, record.TotalCost, record.UpdatedAt.Format(time.RFC3339Nano), record.ID); err != nil {
			return 0, err
		}
		if err := replaceExtraFees(tx, record.ID, record.ExtraFees); err != nil {
			return 0, err
		}
	}
	return len(records), tx.Commit()
}

func (r *BillingSQLiteRepo) BatchImport(records []models.BillingRecord) error {
	if len(records) == 0 {
		return nil
	}
	now := time.Now()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i := range records {
		records[i].ID = 0
		records[i].CreatedAt, records[i].UpdatedAt = now, now
		records[i].CalculateCosts()
		exists, err := billingPeriodExists(tx, records[i].RoomNumber, records[i].BillingMonth, 0)
		if err != nil {
			return err
		}
		if exists {
			return ErrBillingPeriodExists
		}
		id, err := insertBillingRecord(tx, &records[i], false)
		if err != nil {
			return err
		}
		if err := replaceExtraFees(tx, id, records[i].ExtraFees); err != nil {
			return err
		}
		records[i].ID = id
	}
	return tx.Commit()
}

func (r *BillingSQLiteRepo) ExportToJSON(filename string) error {
	data, err := json.MarshalIndent(r.GetAll(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ImportBillingRecordsToSQLite 导入时保留 JSON 的 ID、时间戳、费用和所有额外费用。
// 若目标库已有任何账单则拒绝导入，避免重复迁移造成数据翻倍。
func ImportBillingRecordsToSQLite(db *sql.DB, records []models.BillingRecord) error {
	var existing int
	if err := db.QueryRow(`SELECT COUNT(*) FROM billing_records`).Scan(&existing); err != nil {
		return err
	}
	if existing != 0 {
		return fmt.Errorf("sqlite 已包含 %d 条账单，拒绝重复迁移", existing)
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i := range records {
		if records[i].CreatedAt.IsZero() {
			records[i].CreatedAt = time.Now()
		}
		if records[i].UpdatedAt.IsZero() {
			records[i].UpdatedAt = records[i].CreatedAt
		}
		id, err := insertBillingRecord(tx, &records[i], true)
		if err != nil {
			return fmt.Errorf("导入账单 ID=%d 失败: %w", records[i].ID, err)
		}
		if err := replaceExtraFees(tx, id, records[i].ExtraFees); err != nil {
			return fmt.Errorf("导入账单 ID=%d 的额外费用失败: %w", records[i].ID, err)
		}
	}
	return tx.Commit()
}
