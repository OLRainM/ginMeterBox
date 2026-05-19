package handlers

import (
	"fmt"
	"sort"
	"strconv"

	"ginMeterBox/models"
	"ginMeterBox/pkg/response"
	"ginMeterBox/repository"
	"ginMeterBox/services"

	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	repo         repository.BillingRepo
	imgGenerator *services.ImageGenerator
}

func NewBillingHandler(repo repository.BillingRepo) *BillingHandler {
	return &BillingHandler{
		repo:         repo,
		imgGenerator: services.NewImageGenerator(),
	}
}

// GetAll 获取所有记录
func (h *BillingHandler) GetAll(c *gin.Context) {
	records := h.repo.GetAll()

	sortBy := c.Query("sortBy")
	order := c.Query("order")
	if sortBy == "room" {
		sort.Slice(records, func(i, j int) bool {
			if order == "desc" {
				return records[i].RoomNumber > records[j].RoomNumber
			}
			return records[i].RoomNumber < records[j].RoomNumber
		})
	}

	response.OK(c, records)
}

// GetByID 根据ID获取记录
func (h *BillingHandler) GetByID(c *gin.Context) {
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
	response.OK(c, record)
}

// Create 创建新记录
func (h *BillingHandler) Create(c *gin.Context) {
	var record models.BillingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.Create(&record); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Created(c, record, "")
}

// Update 更新记录
func (h *BillingHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	var record models.BillingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.Update(id, &record); err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "Record not found")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, record)
}

// Delete 删除记录
func (h *BillingHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	if err := h.repo.Delete(id); err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "Record not found")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "Record deleted successfully")
}

// GetByMonth 根据月份获取记录
func (h *BillingHandler) GetByMonth(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		response.BadRequest(c, "Month parameter is required")
		return
	}
	response.OK(c, h.repo.GetByMonth(month))
}

// Calculate 计算费用（不保存）
func (h *BillingHandler) Calculate(c *gin.Context) {
	var record models.BillingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record.CalculateCosts()
	response.OK(c, record)
}

// ContinueFromPrevious 从上月数据自动延续创建新记录
func (h *BillingHandler) ContinueFromPrevious(c *gin.Context) {
	var req struct {
		RoomNumber string `json:"roomNumber" binding:"required"`
		NewMonth   string `json:"newMonth" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record, err := h.createFromPrevious(req.RoomNumber, req.NewMonth)
	if err != nil {
		response.ServerError(c, "自动延续失败: "+err.Error())
		return
	}
	response.Created(c, record, "已从上月数据自动创建新记录")
}

// BatchContinueFromPrevious 批量自动延续
func (h *BillingHandler) BatchContinueFromPrevious(c *gin.Context) {
	var req struct {
		RoomNumbers []string `json:"roomNumbers"`
		NewMonth    string   `json:"newMonth"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(req.RoomNumbers) == 0 {
		response.BadRequest(c, "请选择至少一个住户")
		return
	}
	if req.NewMonth == "" {
		response.BadRequest(c, "新月份不能为空")
		return
	}

	successCount := 0
	failedRooms := []string{}
	for _, roomNumber := range req.RoomNumbers {
		_, err := h.createFromPrevious(roomNumber, req.NewMonth)
		if err != nil {
			failedRooms = append(failedRooms, roomNumber)
		} else {
			successCount++
		}
	}

	if successCount == 0 {
		response.ServerError(c, "批量自动延续失败，所有住户都未能成功创建")
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
	response.OKData(c, result)
}

// GetLatestByRoom 获取某住户的最新记录
func (h *BillingHandler) GetLatestByRoom(c *gin.Context) {
	roomNumber := c.Param("room")
	if roomNumber == "" {
		response.BadRequest(c, "住户编号不能为空")
		return
	}
	record, err := h.repo.GetLatestByRoomNumber(roomNumber)
	if err != nil {
		response.NotFound(c, "未找到该住户的记录")
		return
	}
	response.OK(c, record)
}

// createFromPrevious 从上月记录创建新记录（自动延续）
func (h *BillingHandler) createFromPrevious(roomNumber, newMonth string) (*models.BillingRecord, error) {
	previous, err := h.repo.GetLatestByRoomNumber(roomNumber)
	if err != nil {
		return nil, err
	}

	newRecord := &models.BillingRecord{
		RoomNumber:       roomNumber,
		BillingMonth:     newMonth,
		CurrentWater:     previous.CurrentWater,
		PreviousWater:    previous.CurrentWater,
		CurrentElectric:  previous.CurrentElectric,
		PreviousElectric: previous.CurrentElectric,
		ManagementFee:    previous.ManagementFee,
		WaterPrice:       previous.WaterPrice,
		ElectricPrice:    previous.ElectricPrice,
	}

	if err := h.repo.Create(newRecord); err != nil {
		return nil, err
	}
	return newRecord, nil
}
