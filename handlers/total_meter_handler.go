package handlers

import (
	"ginMeterBox/dto"
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

// GetAll 获取所有总表记录。
func (h *TotalMeterHandler) GetAll(c *gin.Context) {
	response.OK(c, h.repo.GetAll())
}

// GetByMonth 根据月份获取总表记录。
func (h *TotalMeterHandler) GetByMonth(c *gin.Context) {
	month := c.Query("month")
	if err := dto.ValidateMonth(month); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record, err := h.repo.GetByMonth(month)
	if err != nil {
		response.NotFound(c, "未找到该月份的总表记录")
		return
	}
	response.OK(c, record)
}

// Create 创建总表记录。
func (h *TotalMeterHandler) Create(c *gin.Context) {
	var req dto.TotalMeterRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record := req.ToRecord()
	if err := h.repo.Create(&record); err != nil {
		if err == repository.ErrTotalMeterMonthExists {
			response.BadRequest(c, "该月份总表记录已存在")
			return
		}
		response.ServerError(c, "保存总表记录失败")
		return
	}
	response.Created(c, record, "总表记录保存成功")
}

// Update 更新总表记录。
func (h *TotalMeterHandler) Update(c *gin.Context) {
	month := c.Param("month")
	if err := dto.ValidateMonth(month); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var req dto.TotalMeterRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式无效")
		return
	}
	req.Month = month
	if err := req.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record := req.ToRecord()
	if err := h.repo.Update(month, &record); err != nil {
		if err == repository.ErrTotalMeterRecordNotFound {
			response.NotFound(c, "未找到该月份的总表记录")
			return
		}
		response.ServerError(c, "保存总表记录失败")
		return
	}
	response.OK(c, record)
}

// Delete 删除总表记录。
func (h *TotalMeterHandler) Delete(c *gin.Context) {
	month := c.Param("month")
	if err := dto.ValidateMonth(month); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.repo.Delete(month); err != nil {
		if err == repository.ErrTotalMeterRecordNotFound {
			response.NotFound(c, "未找到该月份的总表记录")
			return
		}
		response.ServerError(c, "删除总表记录失败")
		return
	}
	response.OKMsg(c, "总表记录删除成功")
}
