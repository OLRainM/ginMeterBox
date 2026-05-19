package handlers

import (
	"fmt"
	"strings"

	"ginMeterBox/models"
	"ginMeterBox/pkg/response"

	"github.com/gin-gonic/gin"
)

// BatchDelete 批量删除记录
func (h *BillingHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效的请求格式: "+err.Error())
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(c, "请选择要删除的记录")
		return
	}

	successCount := 0
	for _, id := range req.IDs {
		if err := h.repo.Delete(id); err == nil {
			successCount++
		}
	}
	if successCount == 0 {
		response.ServerError(c, "批量删除失败，没有记录被删除")
		return
	}
	response.OKData(c, gin.H{"count": successCount, "message": fmt.Sprintf("成功删除 %d 条记录", successCount)})
}

// BatchSetAdjustment 批量设置水电补差
func (h *BillingHandler) BatchSetAdjustment(c *gin.Context) {
	var request struct {
		IDs                []int    `json:"ids"`
		WaterAdjustment    *float64 `json:"waterAdjustment"`
		ElectricAdjustment *float64 `json:"electricAdjustment"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "无效的请求格式: "+err.Error())
		return
	}
	if len(request.IDs) == 0 {
		response.BadRequest(c, "请选择要设置补差的记录")
		return
	}
	if request.WaterAdjustment == nil && request.ElectricAdjustment == nil {
		response.BadRequest(c, "请至少设置一个补差值")
		return
	}

	count := 0
	var updateErrors []string
	for _, id := range request.IDs {
		record, err := h.repo.GetByID(id)
		if err != nil {
			updateErrors = append(updateErrors, fmt.Sprintf("记录ID %d 不存在", id))
			continue
		}
		if request.WaterAdjustment != nil {
			record.WaterAdjustment = *request.WaterAdjustment
		}
		if request.ElectricAdjustment != nil {
			record.ElectricAdjustment = *request.ElectricAdjustment
		}
		record.CalculateCosts()
		if err := h.repo.Update(id, record); err == nil {
			count++
		} else {
			updateErrors = append(updateErrors, fmt.Sprintf("房号%s更新失败: %v", record.RoomNumber, err))
		}
	}

	if count == 0 {
		errorMsg := "批量设置失败，没有记录被更新"
		if len(updateErrors) > 0 {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, strings.Join(updateErrors, "; "))
		}
		response.ServerError(c, errorMsg)
		return
	}

	result := gin.H{"message": fmt.Sprintf("成功为 %d 条记录设置补差", count), "count": count}
	if len(updateErrors) > 0 {
		result["warnings"] = updateErrors
		result["message"] = fmt.Sprintf("成功为 %d 条记录设置补差，%d 条失败", count, len(updateErrors))
	}
	response.OKData(c, result)
}

// BatchSetExtraFee 批量设置额外费用
func (h *BillingHandler) BatchSetExtraFee(c *gin.Context) {
	var request struct {
		IDs       []int             `json:"ids"`
		ExtraFees []models.ExtraFee `json:"extraFees"`
		Mode      string            `json:"mode"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "无效的请求格式: "+err.Error())
		return
	}
	if len(request.IDs) == 0 {
		response.BadRequest(c, "请选择要设置额外费用的记录")
		return
	}
	if len(request.ExtraFees) == 0 {
		response.BadRequest(c, "请至少添加一项额外费用")
		return
	}
	if request.Mode != "append" && request.Mode != "replace" {
		request.Mode = "append"
	}

	count := 0
	for _, id := range request.IDs {
		record, err := h.repo.GetByID(id)
		if err != nil {
			continue
		}
		if request.Mode == "replace" {
			record.ExtraFees = request.ExtraFees
		} else {
			record.ExtraFees = append(record.ExtraFees, request.ExtraFees...)
		}
		record.CalculateCosts()
		if err := h.repo.Update(id, record); err == nil {
			count++
		}
	}

	if count == 0 {
		response.ServerError(c, "批量设置失败，没有记录被更新")
		return
	}
	response.OKData(c, gin.H{"message": "批量设置成功", "count": count})
}
