package handlers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"ginMeterBox/dto"
	"ginMeterBox/models"
	"ginMeterBox/pkg/response"
	"ginMeterBox/services"

	"github.com/gin-gonic/gin"
)

// GenerateReport 生成账单报表图片
func (h *BillingHandler) GenerateReport(c *gin.Context) {
	var request dto.GenerateReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := request.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var records []models.BillingRecord

	if len(request.IDs) > 0 {
		records = h.service.GetByIDs(request.IDs)
	} else {
		records = h.service.GetByMonth(request.Month)
	}

	if len(records) == 0 {
		response.NotFound(c, "未找到相关记录")
		return
	}

	if request.SortBy == "room" {
		sort.Slice(records, func(i, j int) bool {
			if request.Order == "desc" {
				return records[i].RoomNumber > records[j].RoomNumber
			}
			return records[i].RoomNumber < records[j].RoomNumber
		})
	}

	reportMonth := request.Month
	if reportMonth == "" {
		months := make(map[string]struct{})
		for _, record := range records {
			if record.BillingMonth != "" {
				months[record.BillingMonth] = struct{}{}
			}
		}
		switch len(months) {
		case 1:
			for month := range months {
				reportMonth = month
			}
		case 0:
			reportMonth = "未指定月份"
		default:
			reportMonth = "多月份账单"
		}
	}

	filename, err := h.imgGenerator.GenerateBillingReport(records, reportMonth)
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
