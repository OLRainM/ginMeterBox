package handlers

import (
	"fmt"
	"strings"
	"time"

	"ginMeterBox/models"
	"ginMeterBox/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// BatchImport 批量导入记录
func (h *BillingHandler) BatchImport(c *gin.Context) {
	var records []models.BillingRecord
	if err := c.ShouldBindJSON(&records); err != nil {
		response.BadRequest(c, "无效的JSON格式: "+err.Error())
		return
	}
	if len(records) == 0 {
		response.BadRequest(c, "没有要导入的记录")
		return
	}
	if err := h.repo.BatchImport(records); err != nil {
		response.ServerError(c, "批量导入失败: "+err.Error())
		return
	}
	response.OKData(c, gin.H{"message": "成功导入记录", "count": len(records)})
}

// ExportToJSON 导出所有记录为JSON文件
func (h *BillingHandler) ExportToJSON(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		filename = "exports/billing_export.json"
	}
	if err := h.repo.ExportToJSON(filename); err != nil {
		response.ServerError(c, "导出失败: "+err.Error())
		return
	}
	response.OKData(c, gin.H{"message": "导出成功", "file": filename})
}

// ExportToExcel 导出选中记录为Excel
func (h *BillingHandler) ExportToExcel(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求格式: "+err.Error())
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "请选择要导出的记录")
		return
	}

	var records []models.BillingRecord
	for _, id := range req.IDs {
		if record, err := h.repo.GetByID(id); err == nil {
			records = append(records, *record)
		}
	}
	if len(records) == 0 {
		response.NotFound(c, "未找到要导出的记录")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheetName := "账单记录"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		response.ServerError(c, "创建Excel工作表失败: "+err.Error())
		return
	}

	headers := []string{
		"ID", "住户编号", "缴费月份",
		"上月水表", "本月水表", "水分摊", "用水量", "水单价", "水费",
		"上月电表", "本月电表", "电分摊", "用电量", "电单价", "电费",
		"管理费", "额外费用", "总费用", "创建时间", "更新时间",
	}
	for i, header := range headers {
		f.SetCellValue(sheetName, fmt.Sprintf("%s1", string(rune('A'+i))), header)
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"667EEA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", headerStyle)

	for i, record := range records {
		writeExcelRow(f, sheetName, i+2, record)
	}

	f.SetColWidth(sheetName, "A", "A", 6)
	f.SetColWidth(sheetName, "B", "C", 12)
	f.SetColWidth(sheetName, "D", "O", 10)
	f.SetColWidth(sheetName, "P", "Q", 20)
	f.SetColWidth(sheetName, "R", "R", 10)
	f.SetColWidth(sheetName, "S", "T", 18)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	filename := fmt.Sprintf("exports/billing_export_%s.xlsx", time.Now().Format("20060102150405"))
	if err := f.SaveAs(filename); err != nil {
		response.ServerError(c, "保存Excel文件失败: "+err.Error())
		return
	}
	response.OKData(c, gin.H{
		"message": fmt.Sprintf("成功导出 %d 条记录", len(records)),
		"file":    filename,
		"count":   len(records),
	})
}

func writeExcelRow(f *excelize.File, sheet string, row int, r models.BillingRecord) {
	extraFeesText := ""
	if len(r.ExtraFees) > 0 {
		var fees []string
		for _, fee := range r.ExtraFees {
			fees = append(fees, fmt.Sprintf("%s:¥%.2f", fee.Name, fee.Amount))
		}
		extraFeesText = strings.Join(fees, "; ")
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.ID)
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.RoomNumber)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.BillingMonth)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.PreviousWater)
	f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.CurrentWater)
	f.SetCellValue(sheet, fmt.Sprintf("F%d", row), r.WaterAdjustment)
	f.SetCellValue(sheet, fmt.Sprintf("G%d", row), r.WaterUsage)
	f.SetCellValue(sheet, fmt.Sprintf("H%d", row), r.WaterPrice)
	f.SetCellValue(sheet, fmt.Sprintf("I%d", row), r.TotalWaterCost)
	f.SetCellValue(sheet, fmt.Sprintf("J%d", row), r.PreviousElectric)
	f.SetCellValue(sheet, fmt.Sprintf("K%d", row), r.CurrentElectric)
	f.SetCellValue(sheet, fmt.Sprintf("L%d", row), r.ElectricAdjustment)
	f.SetCellValue(sheet, fmt.Sprintf("M%d", row), r.ElectricUsage)
	f.SetCellValue(sheet, fmt.Sprintf("N%d", row), r.ElectricPrice)
	f.SetCellValue(sheet, fmt.Sprintf("O%d", row), r.TotalElectricCost)
	f.SetCellValue(sheet, fmt.Sprintf("P%d", row), r.ManagementFee)
	f.SetCellValue(sheet, fmt.Sprintf("Q%d", row), extraFeesText)
	f.SetCellValue(sheet, fmt.Sprintf("R%d", row), r.TotalCost)
	f.SetCellValue(sheet, fmt.Sprintf("S%d", row), r.CreatedAt.Format("2006-01-02 15:04:05"))
	f.SetCellValue(sheet, fmt.Sprintf("T%d", row), r.UpdatedAt.Format("2006-01-02 15:04:05"))
}
