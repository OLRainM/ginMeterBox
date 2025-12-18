package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-ele/models"
	"go-ele/services"
	"go-ele/storage"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type BillingHandler struct {
	storage   *storage.Storage
	imgGenerator *services.ImageGenerator
}

func NewBillingHandler(s *storage.Storage) *BillingHandler {
	return &BillingHandler{
		storage:   s,
		imgGenerator: services.NewImageGenerator(),
	}
}

// GetAll 获取所有记录
func (h *BillingHandler) GetAll(c *gin.Context) {
	records := h.storage.GetAll()
	
	// 获取排序参数：sortBy=room&order=asc|desc
	sortBy := c.Query("sortBy")
	order := c.Query("order")
	
	// 如果指定了按房号排序
	if sortBy == "room" {
		sort.Slice(records, func(i, j int) bool {
			if order == "desc" {
				return records[i].RoomNumber > records[j].RoomNumber
			}
			return records[i].RoomNumber < records[j].RoomNumber
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

// GetByID 根据ID获取记录
func (h *BillingHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ID",
		})
		return
	}

	record, err := h.storage.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Record not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// Create 创建新记录
func (h *BillingHandler) Create(c *gin.Context) {
	var record models.BillingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.storage.Create(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    record,
	})
}

// Update 更新记录
func (h *BillingHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ID",
		})
		return
	}

	var record models.BillingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.storage.Update(id, &record); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// Delete 删除记录
func (h *BillingHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ID",
		})
		return
	}

	if err := h.storage.Delete(id); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Record deleted successfully",
	})
}

// GetByMonth 根据月份获取记录
func (h *BillingHandler) GetByMonth(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Month parameter is required",
		})
		return
	}

	records := h.storage.GetByMonth(month)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

// Calculate 计算费用（不保存）
func (h *BillingHandler) Calculate(c *gin.Context) {
	var record models.BillingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 计算费用
	record.CalculateCosts()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// GenerateReport 生成账单报表图片
func (h *BillingHandler) GenerateReport(c *gin.Context) {
	// 获取参数
	idsParam := c.Query("ids")
	month := c.Query("month")
	sortBy := c.Query("sortBy")
	order := c.Query("order")

	var records []models.BillingRecord

	if idsParam != "" {
		// 按ID列表生成
		idStrs := strings.Split(idsParam, ",")
		ids := make([]int, 0, len(idStrs))
		for _, idStr := range idStrs {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err == nil {
				ids = append(ids, id)
			}
		}
		records = h.storage.GetByIDs(ids)
	} else if month != "" {
		// 按月份生成
		records = h.storage.GetByMonth(month)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请提供 ids 或 month 参数",
		})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到相关记录",
		})
		return
	}

	// 按房号排序（如果指定）
	if sortBy == "room" {
		sort.Slice(records, func(i, j int) bool {
			if order == "desc" {
				return records[i].RoomNumber > records[j].RoomNumber
			}
			return records[i].RoomNumber < records[j].RoomNumber
		})
	}

	// 生成图片
	filename, err := h.imgGenerator.GenerateBillingReport(records, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "生成图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"filename": filename,
			"count":    len(records),
			"message":  "报表生成成功",
		},
	})
}

// GenerateCard 生成单个用户的卡片
func (h *BillingHandler) GenerateCard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid ID",
		})
		return
	}

	record, err := h.storage.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Record not found",
		})
		return
	}

	filename, err := h.imgGenerator.GenerateSimpleCard(*record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "生成卡片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"filename": filename,
			"message":  "卡片生成成功",
		},
	})
}

// DownloadImage 下载生成的图片
func (h *BillingHandler) DownloadImage(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "文件名不能为空",
		})
		return
	}

	// 安全检查：确保文件在reports目录下
	if !strings.HasPrefix(filename, "reports/") {
		filename = "reports/" + filename
	}

	c.File(filename)
}

// ContinueFromPrevious 从上月数据自动延续创建新记录
func (h *BillingHandler) ContinueFromPrevious(c *gin.Context) {
	var req struct {
		RoomNumber string `json:"roomNumber" binding:"required"`
		NewMonth   string `json:"newMonth" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	record, err := h.storage.CreateFromPrevious(req.RoomNumber, req.NewMonth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "自动延续失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    record,
		"message": "已从上月数据自动创建新记录",
	})
}

// BatchContinueFromPrevious 批量自动延续
func (h *BillingHandler) BatchContinueFromPrevious(c *gin.Context) {
	var req struct {
		RoomNumbers []string `json:"roomNumbers"`
		NewMonth    string   `json:"newMonth"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(req.RoomNumbers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请选择至少一个住户",
		})
		return
	}

	if req.NewMonth == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "新月份不能为空",
		})
		return
	}

	successCount := 0
	failedRooms := []string{}
	
	for _, roomNumber := range req.RoomNumbers {
		_, err := h.storage.CreateFromPrevious(roomNumber, req.NewMonth)
		if err != nil {
			failedRooms = append(failedRooms, roomNumber)
		} else {
			successCount++
		}
	}

	if successCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量自动延续失败，所有住户都未能成功创建",
			"failed":  failedRooms,
		})
		return
	}

	result := gin.H{
		"success": true,
		"count":   successCount,
		"message": fmt.Sprintf("成功为 %d 个住户创建新记录", successCount),
	}

	if len(failedRooms) > 0 {
		result["partialSuccess"] = true
		result["failed"] = failedRooms
		result["message"] = fmt.Sprintf("成功为 %d 个住户创建新记录，%d 个失败", successCount, len(failedRooms))
	}

	c.JSON(http.StatusCreated, result)
}

// GetLatestByRoom 获取某住户的最新记录
func (h *BillingHandler) GetLatestByRoom(c *gin.Context) {
	roomNumber := c.Param("room")
	if roomNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "住户编号不能为空",
		})
		return
	}

	record, err := h.storage.GetLatestByRoomNumber(roomNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到该住户的记录",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// BatchImport 批量导入记录
func (h *BillingHandler) BatchImport(c *gin.Context) {
	var records []models.BillingRecord
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的JSON格式: " + err.Error(),
		})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "没有要导入的记录",
		})
		return
	}

	if err := h.storage.BatchImport(records); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量导入失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "成功导入记录",
		"count":   len(records),
	})
}

// ExportToJSON 导出所有记录为JSON文件
func (h *BillingHandler) ExportToJSON(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		filename = "exports/billing_export.json"
	}

	if err := h.storage.ExportToJSON(filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "导出失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "导出成功",
		"file":    filename,
	})
}

// ExportToExcel 导出选中记录为Excel
func (h *BillingHandler) ExportToExcel(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求格式: " + err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请选择要导出的记录",
		})
		return
	}

	// 获取选中的记录
	var records []models.BillingRecord
	for _, id := range req.IDs {
		record, err := h.storage.GetByID(id)
		if err == nil {
			records = append(records, *record)
		}
	}

	if len(records) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到要导出的记录",
		})
		return
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "账单记录"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建Excel工作表失败: " + err.Error(),
		})
		return
	}

	// 设置表头
	headers := []string{
		"ID", "住户编号", "缴费月份",
		"上月水表", "本月水表", "水分摊", "用水量", "水单价", "水费",
		"上月电表", "本月电表", "电分摊", "用电量", "电单价", "电费",
		"管理费", "额外费用", "总费用", "创建时间", "更新时间",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"667EEA"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheetName, "A1", string(rune('A'+len(headers)-1))+"1", headerStyle)

	// 写入数据
	for i, record := range records {
		row := i + 2
		
		// 处理额外费用
		extraFeesText := ""
		if len(record.ExtraFees) > 0 {
			var fees []string
			for _, fee := range record.ExtraFees {
				fees = append(fees, fmt.Sprintf("%s:¥%.2f", fee.Name, fee.Amount))
			}
			extraFeesText = strings.Join(fees, "; ")
		}

		// 写入每列数据
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), record.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), record.RoomNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), record.BillingMonth)
		
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), record.PreviousWater)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), record.CurrentWater)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), record.WaterAdjustment)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), record.WaterUsage)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), record.WaterPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), record.TotalWaterCost)
		
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), record.PreviousElectric)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), record.CurrentElectric)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), record.ElectricAdjustment)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), record.ElectricUsage)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), record.ElectricPrice)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), record.TotalElectricCost)
		
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), record.ManagementFee)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), extraFeesText)
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), record.TotalCost)
		f.SetCellValue(sheetName, fmt.Sprintf("S%d", row), record.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, fmt.Sprintf("T%d", row), record.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 6)
	f.SetColWidth(sheetName, "B", "C", 12)
	f.SetColWidth(sheetName, "D", "I", 10)
	f.SetColWidth(sheetName, "J", "O", 10)
	f.SetColWidth(sheetName, "P", "P", 10)
	f.SetColWidth(sheetName, "Q", "Q", 30)
	f.SetColWidth(sheetName, "R", "R", 10)
	f.SetColWidth(sheetName, "S", "T", 18)

	// 设置活动工作表
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 生成文件名
	filename := fmt.Sprintf("exports/billing_export_%s.xlsx", time.Now().Format("20060102150405"))
	
	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存Excel文件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("成功导出 %d 条记录", len(records)),
		"file":    filename,
		"count":   len(records),
	})
}

// BatchDelete 批量删除记录
func (h *BillingHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求格式: " + err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请选择要删除的记录",
		})
		return
	}

	successCount := 0
	for _, id := range req.IDs {
		if err := h.storage.Delete(id); err == nil {
			successCount++
		}
	}

	if successCount == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量删除失败，没有记录被删除",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   successCount,
		"message": fmt.Sprintf("成功删除 %d 条记录", successCount),
	})
}

// BatchSetExtraFee 批量设置额外费用
func (h *BillingHandler) BatchSetExtraFee(c *gin.Context) {
	var request struct {
		IDs       []int             `json:"ids"`
		ExtraFees []models.ExtraFee `json:"extraFees"`
		Mode      string            `json:"mode"` // "append" 或 "replace"
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求格式: " + err.Error(),
		})
		return
	}

	if len(request.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请选择要设置额外费用的记录",
		})
		return
	}

	if len(request.ExtraFees) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请至少添加一项额外费用",
		})
		return
	}

	if request.Mode != "append" && request.Mode != "replace" {
		request.Mode = "append" // 默认为追加模式
	}

	count := 0
	for _, id := range request.IDs {
		record, err := h.storage.GetByID(id)
		if err != nil {
			continue // 跳过不存在的记录
		}

		if request.Mode == "replace" {
			// 替换模式：清空现有额外费用
			record.ExtraFees = request.ExtraFees
		} else {
			// 追加模式：在现有额外费用基础上追加
			record.ExtraFees = append(record.ExtraFees, request.ExtraFees...)
		}

		// 重新计算费用
		record.CalculateCosts()

		// 更新记录
		if err := h.storage.Update(id, record); err == nil {
			count++
		}
	}

	if count == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量设置失败，没有记录被更新",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "批量设置成功",
		"count":   count,
	})
}
