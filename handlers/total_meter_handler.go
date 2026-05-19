package handlers

import (
	"ginMeterBox/models"
	"ginMeterBox/pkg/response"
	"ginMeterBox/repository"

	"github.com/gin-gonic/gin"
)

type TotalMeterHandler struct {
	repo repository.TotalMeterRepo
}

func NewTotalMeterHandler(repo repository.TotalMeterRepo) *TotalMeterHandler {
	return &TotalMeterHandler{repo: repo}
}

// GetAll 获取所有总表记录
func (h *TotalMeterHandler) GetAll(c *gin.Context) {
	response.OK(c, h.repo.GetAll())
}

// GetByMonth 根据月份获取总表记录
func (h *TotalMeterHandler) GetByMonth(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		response.BadRequest(c, "月份参数不能为空")
		return
	}
	record, err := h.repo.GetByMonth(month)
	if err != nil {
		response.NotFound(c, "未找到该月份的总表记录")
		return
	}
	response.OK(c, record)
}

// Create 创建总表记录
func (h *TotalMeterHandler) Create(c *gin.Context) {
	var record models.TotalMeterRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.Create(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, record, "总表记录保存成功")
}

// Update 更新总表记录
func (h *TotalMeterHandler) Update(c *gin.Context) {
	month := c.Param("month")
	var record models.TotalMeterRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record.Month = month
	if err := h.repo.Update(month, &record); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, record)
}

// Delete 删除总表记录
func (h *TotalMeterHandler) Delete(c *gin.Context) {
	month := c.Param("month")
	if err := h.repo.Delete(month); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OKMsg(c, "总表记录删除成功")
}
