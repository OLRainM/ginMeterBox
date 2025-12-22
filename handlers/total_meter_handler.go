package handlers

import (
	"go-ele/models"
	"go-ele/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TotalMeterHandler struct {
	storage *storage.TotalMeterStorage
}

func NewTotalMeterHandler(s *storage.TotalMeterStorage) *TotalMeterHandler {
	return &TotalMeterHandler{storage: s}
}

// GetAll 获取所有总表记录
func (h *TotalMeterHandler) GetAll(c *gin.Context) {
	records := h.storage.GetAll()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

// GetByMonth 根据月份获取总表记录
func (h *TotalMeterHandler) GetByMonth(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "月份参数不能为空",
		})
		return
	}

	record, err := h.storage.GetByMonth(month)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "未找到该月份的总表记录",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

// Create 创建总表记录
func (h *TotalMeterHandler) Create(c *gin.Context) {
	var record models.TotalMeterRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.storage.Create(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    record,
		"message": "总表记录保存成功",
	})
}

// Update 更新总表记录
func (h *TotalMeterHandler) Update(c *gin.Context) {
	month := c.Param("month")
	
	var record models.TotalMeterRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	record.Month = month

	if err := h.storage.Update(month, &record); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
		"message": "总表记录更新成功",
	})
}

// Delete 删除总表记录
func (h *TotalMeterHandler) Delete(c *gin.Context) {
	month := c.Param("month")

	if err := h.storage.Delete(month); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "总表记录删除成功",
	})
}
