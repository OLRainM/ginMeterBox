package handlers

import (
	"fmt"
	"strings"

	"ginMeterBox/models"
	"ginMeterBox/pkg/response"

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
	var request struct {
		IDs           []int     `json:"ids"`
		WaterReadings []float64 `json:"waterReadings"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "无效的请求格式: "+err.Error())
		return
	}
	if len(request.IDs) == 0 {
		response.BadRequest(c, "请选择要匹配的用户")
		return
	}
	if len(request.WaterReadings) == 0 {
		response.BadRequest(c, "请输入水表读数")
		return
	}
	if len(request.IDs) != len(request.WaterReadings) {
		response.BadRequest(c, fmt.Sprintf("用户数量(%d)与水表读数数量(%d)不匹配", len(request.IDs), len(request.WaterReadings)))
		return
	}
	if len(request.IDs) > 10 {
		response.BadRequest(c, "为保证性能，单次匹配用户数量不能超过10个，建议分批处理")
		return
	}

	var records []*models.BillingRecord
	for _, id := range request.IDs {
		record, err := h.repo.GetByID(id)
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

	successCount := 0
	var matchResults []gin.H
	var updateErrors []string
	for _, match := range matches {
		record := match.Record
		record.CurrentWater = match.WaterReading
		record.CalculateCosts()
		if err := h.repo.Update(record.ID, record); err == nil {
			successCount++
			matchResults = append(matchResults, gin.H{
				"id": record.ID, "roomNumber": record.RoomNumber,
				"waterReading": match.WaterReading, "waterUsage": record.WaterUsage,
				"previousWater": record.PreviousWater,
			})
		} else {
			updateErrors = append(updateErrors, fmt.Sprintf("房号%s更新失败: %v", record.RoomNumber, err))
		}
	}

	if successCount == 0 {
		errorMsg := "智能匹配失败，没有记录被更新"
		if len(updateErrors) > 0 {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, strings.Join(updateErrors, "; "))
		}
		response.ServerError(c, errorMsg)
		return
	}

	result := gin.H{"message": fmt.Sprintf("成功匹配并更新 %d 条记录", successCount), "count": successCount, "matches": matchResults}
	if len(updateErrors) > 0 {
		result["warnings"] = updateErrors
		result["message"] = fmt.Sprintf("成功匹配并更新 %d 条记录，%d 条失败", successCount, len(updateErrors))
	}
	response.OKData(c, result)
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
	hasValidMatch := false

	permutations := generatePermutations(readings)
	for _, perm := range permutations {
		totalUsage := 0.0
		currentMatches := make([]WaterMatch, n)
		isValid := true
		for i := 0; i < n; i++ {
			usage := perm[i] - records[i].PreviousWater + records[i].WaterAdjustment
			if usage < 0 {
				isValid = false
				break
			}
			totalUsage += usage
			currentMatches[i] = WaterMatch{Record: records[i], WaterReading: perm[i], WaterUsage: usage}
		}
		if isValid && totalUsage < minTotalUsage {
			minTotalUsage = totalUsage
			hasValidMatch = true
			bestMatches = make([]WaterMatch, n)
			copy(bestMatches, currentMatches)
		}
	}

	if !hasValidMatch {
		return []WaterMatch{}
	}
	return bestMatches
}

// generatePermutations 生成所有排列组合
func generatePermutations(arr []float64) [][]float64 {
	var result [][]float64
	n := len(arr)
	if n == 0 {
		return result
	}
	if n == 1 {
		return [][]float64{{arr[0]}}
	}
	var backtrack func([]float64, int)
	backtrack = func(current []float64, start int) {
		if start == n {
			perm := make([]float64, n)
			copy(perm, current)
			result = append(result, perm)
			return
		}
		for i := start; i < n; i++ {
			current[start], current[i] = current[i], current[start]
			backtrack(current, start+1)
			current[start], current[i] = current[i], current[start]
		}
	}
	working := make([]float64, n)
	copy(working, arr)
	backtrack(working, 0)
	return result
}
