package services

import "ginMeterBox/models"

// MatchResult 匹配结果
type MatchResult struct {
	Record       *models.BillingRecord
	WaterReading float64
	WaterUsage   float64
}

// MatchService 智能匹配服务
type MatchService struct{}

func NewMatchService() *MatchService {
	return &MatchService{}
}

// SmartMatch 智能匹配水表读数（最小总用水量原则，且所有用水量必须非负）
func (s *MatchService) SmartMatch(records []*models.BillingRecord, readings []float64) []MatchResult {
	n := len(records)
	if n == 0 {
		return nil
	}
	if n == 1 {
		usage := readings[0] - records[0].PreviousWater + records[0].WaterAdjustment
		return []MatchResult{{Record: records[0], WaterReading: readings[0], WaterUsage: usage}}
	}

	var bestMatches []MatchResult
	minTotalUsage := float64(1e18)
	hasValid := false

	for _, perm := range permutations(readings) {
		totalUsage := 0.0
		current := make([]MatchResult, n)
		valid := true
		for i := 0; i < n; i++ {
			usage := perm[i] - records[i].PreviousWater + records[i].WaterAdjustment
			if usage < 0 {
				valid = false
				break
			}
			totalUsage += usage
			current[i] = MatchResult{Record: records[i], WaterReading: perm[i], WaterUsage: usage}
		}
		if valid && totalUsage < minTotalUsage {
			minTotalUsage = totalUsage
			hasValid = true
			bestMatches = make([]MatchResult, n)
			copy(bestMatches, current)
		}
	}

	if !hasValid {
		return nil
	}
	return bestMatches
}

func permutations(arr []float64) [][]float64 {
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
