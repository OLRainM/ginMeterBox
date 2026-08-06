package handlers

import (
	"ginMeterBox/dto"
	"ginMeterBox/pkg/response"
	"ginMeterBox/repository"

	"github.com/gin-gonic/gin"
)

// BatchDelete 批量删除记录。
func (h *BillingHandler) BatchDelete(c *gin.Context) {
	var req dto.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	count, err := h.service.BatchDelete(req.IDs)
	if err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "存在未找到的记录，未执行删除")
			return
		}
		response.ServerError(c, "批量删除失败")
		return
	}
	response.OK(c, gin.H{"count": count, "message": "批量删除成功"})
}

// BatchSetAdjustment 批量设置水电补差。
func (h *BillingHandler) BatchSetAdjustment(c *gin.Context) {
	var req dto.BatchAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	count, err := h.service.BatchUpdateAdjustments(req.IDs, req.WaterAdjustment, req.ElectricAdjustment)
	if err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "存在未找到的记录，未执行更新")
			return
		}
		response.ServerError(c, "批量设置补差失败")
		return
	}
	response.OK(c, gin.H{"count": count, "message": "批量设置补差成功"})
}

// BatchSetExtraFee 批量设置额外费用。
func (h *BillingHandler) BatchSetExtraFee(c *gin.Context) {
	var req dto.BatchExtraFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Mode == "" {
		req.Mode = "append"
	}
	count, err := h.service.BatchSetExtraFees(req.IDs, req.ExtraFees, req.Mode)
	if err != nil {
		if err == repository.ErrRecordNotFound {
			response.NotFound(c, "存在未找到的记录，未执行更新")
			return
		}
		response.ServerError(c, "批量设置额外费用失败")
		return
	}
	response.OK(c, gin.H{"count": count, "message": "批量设置额外费用成功"})
}
