package handlers

import (
	"fmt"

	"ginMeterBox/dto"
	"ginMeterBox/models"
	"ginMeterBox/pkg/response"
	"ginMeterBox/repository"

	"github.com/gin-gonic/gin"
)

// WaterMatch 水表匹配结果
type WaterMatch struct {
	Record       *models.BillingRecord
	WaterReading float64
	WaterUsage   float64
}

// SmartWaterMatch 智能水表匹配
func (h *BillingHandler) SmartWaterMatch(c *gin.Context) {
	var request dto.SmartWaterMatchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := request.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var records []*models.BillingRecord
	for _, id := range request.IDs {
		record, err := h.service.GetByID(id)
		if err != nil {
			response.NotFound(c, fmt.Sprintf("记录ID %d 不存在", id))
			return
		}
		records = append(records, record)
	}

	matches := smartMatchWaterReadings(records, request.WaterReadings)
	if len(matches) == 0 {
		response.BadRequest(c, "未找到有效的匹配方案：所有可能的匹配都会产生负数用水量。请检查输入的水表读数是否正确。")
		return
	}

	updates := make([]repository.WaterReadingUpdate, len(matches))
	matchResults := make([]gin.H, 0, len(matches))
	for i, match := range matches {
		updates[i] = repository.WaterReadingUpdate{ID: match.Record.ID, CurrentWater: match.WaterReading}
		matchResults = append(matchResults, gin.H{
			"id": match.Record.ID, "roomNumber": match.Record.RoomNumber,
			"waterReading": match.WaterReading, "waterUsage": match.WaterUsage,
			"previousWater": match.Record.PreviousWater,
		})
	}
	if err := h.service.BatchUpdateWaterReadings(updates); err != nil {
		response.ServerError(c, "智能匹配失败，未更新任何记录")
		return
	}
	response.OKData(c, gin.H{"message": fmt.Sprintf("成功匹配并更新 %d 条记录", len(matches)), "count": len(matches), "matches": matchResults})
}

// smartMatchWaterReadings 智能匹配水表读数（最小总用水量原则，且所有用水量必须非负）
func smartMatchWaterReadings(records []*models.BillingRecord, readings []float64) []WaterMatch {
	n := len(records)
	if n == 0 {
		return []WaterMatch{}
	}
	if n == 1 {
		usage := readings[0] - records[0].PreviousWater + records[0].WaterAdjustment
		return []WaterMatch{{Record: records[0], WaterReading: readings[0], WaterUsage: usage}}
	}

	var bestMatches []WaterMatch
	minTotalUsage := float64(1e18)
	used := make([]bool, n)
	currentMatches := make([]WaterMatch, n)

	// 回溯时直接评估候选分配，不再预先分配并保存 n! 个排列，显著降低内存占用。
	var backtrack func(int, float64)
	backtrack = func(index int, totalUsage float64) {
		if totalUsage >= minTotalUsage {
			return
		}
		if index == n {
			minTotalUsage = totalUsage
			bestMatches = append([]WaterMatch(nil), currentMatches...)
			return
		}
		for readingIndex, reading := range readings {
			if used[readingIndex] {
				continue
			}
			usage := reading - records[index].PreviousWater + records[index].WaterAdjustment
			if usage < 0 {
				continue
			}
			used[readingIndex] = true
			currentMatches[index] = WaterMatch{Record: records[index], WaterReading: reading, WaterUsage: usage}
			backtrack(index+1, totalUsage+usage)
			used[readingIndex] = false
		}
	}
	backtrack(0, 0)
	return bestMatches
}
