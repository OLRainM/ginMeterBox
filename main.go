package main

import (
	"log"
	"net/http"

	"go-ele/handlers"
	"go-ele/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建存储实例
	store := storage.NewStorage()
	totalMeterStore := storage.NewTotalMeterStorage()

	// 创建处理器
	billingHandler := handlers.NewBillingHandler(store)
	totalMeterHandler := handlers.NewTotalMeterHandler(totalMeterStore)

	// 创建Gin路由
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/reports", "./reports")
	r.Static("/exports", "./exports")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/total-meter.html", "./static/total-meter.html")
	r.StaticFile("/debug.html", "./static/debug.html")

	// API路由
	api := r.Group("/api/v1")
	{
		// 账单相关路由
		billing := api.Group("/billing")
		{
			billing.GET("", billingHandler.GetAll)           // 获取所有记录
			billing.GET("/:id", billingHandler.GetByID)      // 根据ID获取
			billing.POST("", billingHandler.Create)          // 创建新记录
			billing.PUT("/:id", billingHandler.Update)       // 更新记录
			billing.DELETE("/:id", billingHandler.Delete)    // 删除记录
			billing.GET("/month", billingHandler.GetByMonth) // 按月份查询
			billing.POST("/calculate", billingHandler.Calculate) // 计算费用
			
			// 新功能：图片生成
			billing.GET("/report/generate", billingHandler.GenerateReport)     // 生成报表图片
			billing.GET("/card/:id", billingHandler.GenerateCard)              // 生成单个卡片
			billing.GET("/download", billingHandler.DownloadImage)             // 下载图片
			
			// 新功能：自动延续
			billing.POST("/continue", billingHandler.ContinueFromPrevious)     // 从上月数据创建
			billing.POST("/batch-continue", billingHandler.BatchContinueFromPrevious) // 批量自动延续
			billing.GET("/latest/:room", billingHandler.GetLatestByRoom)       // 获取最新记录
			
			// 新功能：批量导入导出
			billing.POST("/import", billingHandler.BatchImport)                // 批量导入JSON
			billing.GET("/export", billingHandler.ExportToJSON)                // 导出为JSON
			billing.POST("/export-excel", billingHandler.ExportToExcel)        // 导出选中记录为Excel
			
			// 新功能：批量设置额外费用
			billing.POST("/batch-extra-fee", billingHandler.BatchSetExtraFee)  // 批量设置额外费用
			
			// 新功能：批量设置补差
			billing.POST("/batch-adjustment", billingHandler.BatchSetAdjustment) // 批量设置水电补差
			
			// 新功能：批量删除
			billing.POST("/batch-delete", billingHandler.BatchDelete)          // 批量删除记录
			
			// 新功能：智能水表匹配
			billing.POST("/smart-water-match", billingHandler.SmartWaterMatch) // 智能水表匹配
		}

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"message": "Water and Electric Billing System is running",
			})
		})

		// 总表管理路由
		totalMeter := api.Group("/total-meter")
		{
			totalMeter.GET("", totalMeterHandler.GetAll)           // 获取所有总表记录
			totalMeter.GET("/month", totalMeterHandler.GetByMonth) // 根据月份获取
			totalMeter.POST("", totalMeterHandler.Create)          // 创建总表记录
			totalMeter.PUT("/:month", totalMeterHandler.Update)    // 更新总表记录
			totalMeter.DELETE("/:month", totalMeterHandler.Delete) // 删除总表记录
		}
	}

	// 启动服务器
	log.Println("Server starting on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
