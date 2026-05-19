package handlers

import (
	"sort"
	"strconv"
	"strings"

	"ginMeterBox/models"
	"ginMeterBox/pkg/response"

	"github.com/gin-gonic/gin"
)

// GenerateReport 生成账单报表图片
func (h *BillingHandler) GenerateReport(c *gin.Context) {
	idsParam := c.Query("ids")
	month := c.Query("month")
	sortBy := c.Query("sortBy")
	order := c.Query("order")

	var records []models.BillingRecord

	if idsParam != "" {
		idStrs := strings.Split(idsParam, ",")
		ids := make([]int, 0, len(idStrs))
		for _, idStr := range idStrs {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err == nil {
				ids = append(ids, id)
			}
		}
		records = h.repo.GetByIDs(ids)
	} else if month != "" {
		records = h.repo.GetByMonth(month)
	} else {
		response.BadRequest(c, "请提供 ids 或 month 参数")
		return
	}

	if len(records) == 0 {
		response.NotFound(c, "未找到相关记录")
		return
	}

	if sortBy == "room" {
		sort.Slice(records, func(i, j int) bool {
			if order == "desc" {
				return records[i].RoomNumber > records[j].RoomNumber
			}
			return records[i].RoomNumber < records[j].RoomNumber
		})
	}

	filename, err := h.imgGenerator.GenerateBillingReport(records, month)
	if err != nil {
		response.ServerError(c, "生成图片失败: "+err.Error())
		return
	}

	response.OKData(c, gin.H{
		"data": gin.H{"filename": filename, "count": len(records), "message": "报表生成成功"},
	})
}

// GenerateCard 生成单个用户的卡片
func (h *BillingHandler) GenerateCard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	record, err := h.repo.GetByID(id)
	if err != nil {
		response.NotFound(c, "Record not found")
		return
	}
	filename, err := h.imgGenerator.GenerateSimpleCard(*record)
	if err != nil {
		response.ServerError(c, "生成卡片失败: "+err.Error())
		return
	}
	response.OKData(c, gin.H{
		"data": gin.H{"filename": filename, "message": "卡片生成成功"},
	})
}

// DownloadImage 下载生成的图片
func (h *BillingHandler) DownloadImage(c *gin.Context) {
	filename := c.Query("file")
	if filename == "" {
		response.BadRequest(c, "文件名不能为空")
		return
	}
	if !strings.HasPrefix(filename, "reports/") {
		filename = "reports/" + filename
	}
	c.File(filename)
}
