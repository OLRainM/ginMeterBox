// 用法：go run ./cmd/migrate
// 该一次性工具以当前 config.json 为准，将 JSON 账单导入空 SQLite 数据库。
// 应用正常启动时也会自动执行同样的幂等迁移；保留此工具是为了在停服维护窗口显式完成迁移。
package main

import (
	"fmt"
	"log"

	"ginMeterBox/internal/config"
	"ginMeterBox/internal/repository"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatal(err)
	}
	db, err := repository.OpenSQLite(cfg.Data.DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := repository.MigrateJSONToSQLiteIfNeeded(db, cfg.Data.BillingFile, cfg.Data.TotalMeterFile); err != nil {
		log.Fatal(err)
	}
	backfilled, err := repository.BackfillLegacyMasterBillsToTotalMeters(db)
	if err != nil {
		log.Fatal(err)
	}
	var billingCount, totalMeterCount, historicalMasterCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM billing_records`).Scan(&billingCount); err != nil {
		log.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM total_meter_records`).Scan(&totalMeterCount); err != nil {
		log.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM billing_records WHERE room_number = '总表'`).Scan(&historicalMasterCount); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SQLite 迁移完成：账单 %d 条，历史总表账单 %d 条，独立总表记录 %d 条，本次回填 %d 条\n", billingCount, historicalMasterCount, totalMeterCount, backfilled)
}
