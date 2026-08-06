package handlers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"ginMeterBox/models"
	"ginMeterBox/pkg/response"
	"ginMeterBox/services"

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
		records = h.service.GetByIDs(ids)
	} else if month != "" {
		records = h.service.GetByMonth(month)
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
		response.ServerError(c, "生成图片失败")
		return
	}

	response.OKData(c, gin.H{
		"count":       len(records),
		"message":     "报表生成成功",
		"downloadUrl": h.fileStore.ReportDownloadURL(filename),
		"openUrl":     h.fileStore.ReportDownloadURL(filename),
	})
}

// GenerateCard 生成单个用户的卡片
func (h *BillingHandler) GenerateCard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	record, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "Record not found")
		return
	}
	filename, err := h.imgGenerator.GenerateSimpleCard(*record)
	if err != nil {
		response.ServerError(c, "生成卡片失败")
		return
	}
	response.OKData(c, gin.H{
		"message":     "卡片生成成功",
		"downloadUrl": h.fileStore.ReportDownloadURL(filename),
		"openUrl":     h.fileStore.ReportDownloadURL(filename),
	})
}

// DownloadImage 仅下载 reportDir 中由服务端生成的 PNG 文件。
func (h *BillingHandler) DownloadImage(c *gin.Context) {
	filename := c.Query("file")
	path, err := h.fileStore.ResolveReportDownload(filename)
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
	c.Header("Content-Disposition", "inline")
	http.ServeFile(c.Writer, c.Request, path)
}
