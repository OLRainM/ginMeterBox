package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ginMeterBox/dto"
	"ginMeterBox/models"
	"ginMeterBox/pkg/response"
	"ginMeterBox/repository"
	"ginMeterBox/services"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// BatchImport 批量导入记录。
func (h *BillingHandler) BatchImport(c *gin.Context) {
	var requests []dto.BillingRecordRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if len(requests) == 0 || len(requests) > dto.MaxBatchSize {
		response.BadRequest(c, fmt.Sprintf("导入记录数量必须在 1 到 %d 之间", dto.MaxBatchSize))
		return
	}

	records := make([]models.BillingRecord, len(requests))
	for i, request := range requests {
		if err := request.Validate(); err != nil {
			response.BadRequest(c, fmt.Sprintf("第 %d 条记录无效：%s", i+1, err.Error()))
			return
		}
		records[i] = request.ToRecord()
	}
	if err := h.service.BatchImport(records); err != nil {
		if err == repository.ErrBillingPeriodExists {
			response.BadRequest(c, "导入数据包含已存在或重复的住户月份账单")
			return
		}
		response.ServerError(c, "批量导入失败")
		return
	}
	response.OKData(c, gin.H{"message": "成功导入记录", "count": len(records)})
}

// ExportToJSON 导出所有记录；文件名和目录均由服务器固定生成。
func (h *BillingHandler) ExportToJSON(c *gin.Context) {
	basename, filename, err := h.fileStore.NewExportFile(".json")
	if err != nil {
		response.ServerError(c, "创建导出文件失败")
		return
	}
	if err := h.service.ExportToJSON(filename); err != nil {
		response.ServerError(c, "导出失败")
		return
	}
	response.OKData(c, gin.H{
		"message":     "导出成功",
		"downloadUrl": h.fileStore.ExportDownloadURL(basename),
		"openUrl":     h.fileStore.ExportDownloadURL(basename),
	})
}

// DownloadExport 仅下载 exportDir 中由服务端生成的 JSON 或 Excel 文件。
func (h *BillingHandler) DownloadExport(c *gin.Context) {
	path, err := h.fileStore.ResolveExportDownload(c.Query("file"))
	if err != nil {
		if errors.Is(err, services.ErrInvalidGeneratedFile) {
			response.BadRequest(c, "文件标识无效")
			return
		}
		if errors.Is(err, services.ErrGeneratedFileNotFound) {
			response.NotFound(c, "文件不存在")
			return
		}
		response.ServerError(c, "读取文件失败")
		return
	}
	c.Header("Content-Disposition", "attachment")
	http.ServeFile(c.Writer, c.Request, path)
}

// ExportToExcel 导出选中记录为 Excel。
func (h *BillingHandler) ExportToExcel(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	records := h.service.GetByIDs(req.IDs)
	if len(records) != len(req.IDs) {
		response.NotFound(c, "存在未找到的记录")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheetName := "账单记录"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		response.ServerError(c, "创建 Excel 工作表失败")
		return
	}

	headers := []string{
		"ID", "住户编号", "缴费月份",
		"上月水表", "本月水表", "水分摊", "用水量", "水单价", "水费",
		"上月电表", "本月电表", "电分摊", "用电量", "电单价", "电费",
		"管理费", "额外费用", "总费用", "创建时间", "更新时间",
	}
	for i, header := range headers {
		f.SetCellValue(sheetName, fmt.Sprintf("%s1", string(rune('A'+i))), excelSafeText(header))
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

	basename, filename, err := h.fileStore.NewExportFile(".xlsx")
	if err != nil {
		response.ServerError(c, "创建导出文件失败")
		return
	}
	if err := f.SaveAs(filename); err != nil {
		response.ServerError(c, "保存 Excel 文件失败")
		return
	}
	response.OKData(c, gin.H{
		"message":     fmt.Sprintf("成功导出 %d 条记录", len(records)),
		"count":       len(records),
		"downloadUrl": h.fileStore.ExportDownloadURL(basename),
		"openUrl":     h.fileStore.ExportDownloadURL(basename),
	})
}

func writeExcelRow(f *excelize.File, sheet string, row int, r models.BillingRecord) {
	extraFeesText := ""
	if len(r.ExtraFees) > 0 {
		fees := make([]string, 0, len(r.ExtraFees))
		for _, fee := range r.ExtraFees {
			fees = append(fees, fmt.Sprintf("%s:¥%.2f", fee.Name, fee.Amount))
		}
		extraFeesText = strings.Join(fees, "; ")
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), r.ID)
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), excelSafeText(r.RoomNumber))
	f.SetCellValue(sheet, fmt.Sprintf("C%d", row), excelSafeText(r.BillingMonth))
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
	f.SetCellValue(sheet, fmt.Sprintf("Q%d", row), excelSafeText(extraFeesText))
	f.SetCellValue(sheet, fmt.Sprintf("R%d", row), r.TotalCost)
	f.SetCellValue(sheet, fmt.Sprintf("S%d", row), excelSafeText(r.CreatedAt.Format("2006-01-02 15:04:05")))
	f.SetCellValue(sheet, fmt.Sprintf("T%d", row), excelSafeText(r.UpdatedAt.Format("2006-01-02 15:04:05")))
}

func excelSafeText(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
