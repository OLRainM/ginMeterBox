package handlers

import (
	"errors"
	"sort"
	"strconv"

	"ginMeterBox/dto"
	"ginMeterBox/pkg/response"
	"ginMeterBox/repository"
	"ginMeterBox/services"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	service      *services.BillingService
	imgGenerator *services.ImageGenerator
	fileStore    *services.GeneratedFileStore
}

func NewBillingHandler(service *services.BillingService, imgGenerator *services.ImageGenerator, fileStore *services.GeneratedFileStore) *BillingHandler {
	return &BillingHandler{service: service, imgGenerator: imgGenerator, fileStore: fileStore}
}

// GetAll 获取所有记录。
func (h *BillingHandler) GetAll(c *gin.Context) {
	records := h.service.GetAll()
	if c.Query("sortBy") == "room" {
		sort.Slice(records, func(i, j int) bool {
			if c.Query("order") == "desc" {
				return records[i].RoomNumber > records[j].RoomNumber
			}
			return records[i].RoomNumber < records[j].RoomNumber
		})
	}
	pageParam, pageSizeParam := c.Query("page"), c.Query("pageSize")
	if pageParam == "" && pageSizeParam == "" {
		response.OK(c, records)
		return
	}

	page, pageSize, err := parsePagination(pageParam, pageSizeParam)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	total := len(records)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	response.OK(c, gin.H{
		"items":    records[start:end],
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
	})
}

func parsePagination(pageParam, pageSizeParam string) (int, int, error) {
	page, pageSize := 1, 20
	var err error
	if pageParam != "" {
		page, err = strconv.Atoi(pageParam)
		if err != nil || page < 1 {
			return 0, 0, errors.New("page 必须为正整数")
		}
	}
	if pageSizeParam != "" {
		pageSize, err = strconv.Atoi(pageSizeParam)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return 0, 0, errors.New("pageSize 必须为 1 到 100 之间的整数")
		}
	}
	return page, pageSize, nil
}

// GetByID 根据 ID 获取记录。
func (h *BillingHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "记录 ID 无效")
		return
	}
	record, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(c, "记录不存在")
		return
	}
	response.OK(c, record)
}

// Create 创建新记录。
func (h *BillingHandler) Create(c *gin.Context) {
	var req dto.BillingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record := req.ToRecord()
	if err := h.service.Create(&record); err != nil {
		response.ServerError(c, "保存记录失败")
		return
	}
	response.Created(c, record, "")
}

// Update 更新记录。
func (h *BillingHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "记录 ID 无效")
		return
	}
	var req dto.BillingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record := req.ToRecord()
	if err := h.service.Update(id, &record); err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "记录不存在")
			return
		}
		response.ServerError(c, "保存记录失败")
		return
	}
	response.OK(c, record)
}

// Delete 删除记录。
func (h *BillingHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.BadRequest(c, "记录 ID 无效")
		return
	}
	if err := h.service.Delete(id); err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "记录不存在")
			return
		}
		response.ServerError(c, "删除记录失败")
		return
	}
	response.OKMsg(c, "记录已删除")
}

// GetByMonth 根据月份获取记录。
func (h *BillingHandler) GetByMonth(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		response.BadRequest(c, "月份参数不能为空")
		return
	}
	response.OK(c, h.service.GetByMonth(month))
}

// Calculate 计算费用（不保存）。
func (h *BillingHandler) Calculate(c *gin.Context) {
	var req dto.BillingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record := req.ToRecord()
	record.CalculateCosts()
	response.OK(c, record)
}

// ContinueFromPrevious 从上月数据自动延续创建新记录。
func (h *BillingHandler) ContinueFromPrevious(c *gin.Context) {
	var req dto.ContinueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record, err := h.service.ContinueFromPrevious(req.RoomNumber, req.NewMonth)
	if err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "未找到该住户的历史记录")
			return
		}
		response.ServerError(c, "自动延续失败")
		return
	}
	response.Created(c, record, "已从上月数据自动创建新记录")
}

// BatchContinueFromPrevious 批量自动延续。
func (h *BillingHandler) BatchContinueFromPrevious(c *gin.Context) {
	var req dto.BatchContinueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	successCount := 0
	failedRooms := []string{}
	for _, roomNumber := range req.RoomNumbers {
		if _, err := h.service.ContinueFromPrevious(roomNumber, req.NewMonth); err != nil {
			failedRooms = append(failedRooms, roomNumber)
			continue
		}
		successCount++
	}
	if successCount == 0 {
		response.ServerError(c, "批量自动延续失败")
		return
	}
	response.OK(c, gin.H{"count": successCount, "failed": failedRooms})
}

// GetLatestByRoom 获取某住户的最新记录。
func (h *BillingHandler) GetLatestByRoom(c *gin.Context) {
	roomNumber := c.Param("room")
	if roomNumber == "" {
		response.BadRequest(c, "住户编号不能为空")
		return
	}
	record, err := h.service.GetLatestByRoom(roomNumber)
	if err != nil {
		response.NotFound(c, "未找到该住户的记录")
		return
	}
	response.OK(c, record)
}
