package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ginMeterBox/authentication"
	"ginMeterBox/config"
	"ginMeterBox/handlers"
	"ginMeterBox/repository"
	"ginMeterBox/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	if strings.TrimSpace(cfg.Security.AdminPasswordHash) == "" {
		log.Fatal("安全配置无效: security.adminPasswordHash 不能为空")
	}
	if _, err := bcrypt.Cost([]byte(cfg.Security.AdminPasswordHash)); err != nil {
		log.Fatal("安全配置无效: security.adminPasswordHash 不是有效的 bcrypt 哈希")
	}

	// SQLite 是唯一运行时数据源。首次启动时仅在空数据库中从 JSON 事务化导入，
	// 旧 JSON 保留为备份与回滚依据，后续读写不会再修改它。
	database, err := repository.OpenSQLite(cfg.Data.DatabaseFile)
	if err != nil {
		log.Fatal("打开 SQLite 数据库失败:", err)
	}
	defer database.Close()
	if err := repository.MigrateJSONToSQLiteIfNeeded(database, cfg.Data.BillingFile, cfg.Data.TotalMeterFile); err != nil {
		log.Fatal("迁移 JSON 数据到 SQLite 失败:", err)
	}
	backfilled, err := repository.BackfillLegacyMasterBillsToTotalMeters(database)
	if err != nil {
		log.Fatal("回填历史总表读数失败:", err)
	}
	if backfilled > 0 {
		log.Printf("已从历史总表账单回填 %d 条独立总表读数", backfilled)
	}
	billingRepo := repository.NewBillingSQLiteRepo(database)
	totalMeterRepo := repository.NewTotalMeterSQLiteRepo(database)

	// 创建处理器
	fileStore := services.NewGeneratedFileStore(cfg.Export.Dir, cfg.Report.Dir)
	imageGenerator := services.NewImageGenerator(services.ImageGeneratorOptions{
		FileStore:   fileStore,
		BoldFont:    cfg.Font.Bold,
		RegularFont: cfg.Font.Regular,
	})
	billingHandler := handlers.NewBillingHandler(services.NewBillingService(billingRepo), imageGenerator, fileStore)
	totalMeterHandler := handlers.NewTotalMeterHandler(totalMeterRepo)

	// 创建Gin路由
	r := gin.Default()

	// 为每个请求附加可追踪的 ID，并记录不含请求体的结构化访问日志。
	r.Use(requestIDAndAccessLog())

	// 限制请求体，避免大请求占用过多内存。
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})
	r.Use(securityHeaders())

	// 空白名单表示仅同源访问；此时不注册 CORS 中间件，避免空 AllowOrigins 被
	// gin-contrib/cors 判定为冲突配置。静态页面与 API 同源时本来就不需要 CORS。
	if len(cfg.Security.AllowedOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.Security.AllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
			ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
			AllowCredentials: true,
		}))
	}

	// 静态文件服务。调试页不应作为生产静态资源暴露。
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/debug.html" || c.Request.URL.Path == "/static/debug.html" {
			c.Status(http.StatusNotFound)
			c.Abort()
			return
		}
		c.Next()
	})
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/total-meter.html", "./static/total-meter.html")

	// API路由
	authService := authentication.NewService(cfg.Security.AdminPasswordHash, cfg.Security.SessionCookieSecure)
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", validateWriteOrigin(cfg.Security.AllowedOrigins), authService.Login)
			auth.POST("/logout", authService.RequireAuth(), validateWriteOrigin(cfg.Security.AllowedOrigins), authService.Logout)
			auth.GET("/session", authService.RequireAuth(), authService.Session)
		}

		// 账单相关路由
		billing := api.Group("/billing", authService.RequireAuth(), validateWriteOrigin(cfg.Security.AllowedOrigins))
		{
			billing.GET("", billingHandler.GetAll)               // 获取所有记录
			billing.GET("/:id", billingHandler.GetByID)          // 根据ID获取
			billing.POST("", billingHandler.Create)              // 创建新记录
			billing.PUT("/:id", billingHandler.Update)           // 更新记录
			billing.DELETE("/:id", billingHandler.Delete)        // 删除记录
			billing.GET("/month", billingHandler.GetByMonth)     // 按月份查询
			billing.POST("/calculate", billingHandler.Calculate) // 计算费用

			// 新功能：图片生成
			billing.POST("/report/generate", billingHandler.GenerateReport) // 生成报表图片
			billing.POST("/card/:id", billingHandler.GenerateCard)          // 生成单个卡片
			billing.GET("/download", billingHandler.DownloadImage)          // 下载图片

			// 新功能：自动延续
			billing.POST("/continue", billingHandler.ContinueFromPrevious)            // 从上月数据创建
			billing.POST("/batch-continue", billingHandler.BatchContinueFromPrevious) // 批量自动延续
			billing.GET("/latest/:room", billingHandler.GetLatestByRoom)              // 获取最新记录

			// 新功能：批量导入导出
			billing.POST("/import", billingHandler.BatchImport)            // 批量导入JSON
			billing.GET("/export", billingHandler.ExportToJSON)            // 导出为JSON
			billing.GET("/export/download", billingHandler.DownloadExport) // 受限下载导出文件
			billing.POST("/export-excel", billingHandler.ExportToExcel)    // 导出选中记录为Excel

			// 新功能：批量设置额外费用
			billing.POST("/batch-extra-fee", billingHandler.BatchSetExtraFee) // 批量设置额外费用

			// 新功能：批量设置补差
			billing.POST("/batch-adjustment", billingHandler.BatchSetAdjustment) // 批量设置水电补差

			// 新功能：批量删除
			billing.POST("/batch-delete", billingHandler.BatchDelete) // 批量删除记录

			// 新功能：智能水表匹配
			billing.POST("/smart-water-match", billingHandler.SmartWaterMatch) // 智能水表匹配
		}

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"message": "Water and Electric Billing System is running",
			})
		})

		// 总表管理路由
		totalMeter := api.Group("/total-meter", authService.RequireAuth(), validateWriteOrigin(cfg.Security.AllowedOrigins))
		{
			totalMeter.GET("", totalMeterHandler.GetAll)           // 获取所有总表记录
			totalMeter.GET("/month", totalMeterHandler.GetByMonth) // 根据月份获取
			totalMeter.POST("", totalMeterHandler.Create)          // 创建总表记录
			totalMeter.PUT("/:month", totalMeterHandler.Update)    // 更新总表记录
			totalMeter.DELETE("/:month", totalMeterHandler.Delete) // 删除总表记录
		}
	}

	server := &http.Server{
		Addr:              cfg.Server.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on http://localhost%s\n", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server:", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("服务器优雅关闭失败: %v", err)
	}
}
