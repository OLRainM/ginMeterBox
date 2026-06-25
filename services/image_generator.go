package services

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"

	"ginMeterBox/models"

	"github.com/fogleman/gg"
)

// ImageGenerator 图片生成器
type ImageGenerator struct {
	width  int
	height int
}

// NewImageGenerator 创建图片生成器
func NewImageGenerator() *ImageGenerator {
	return &ImageGenerator{
		width:  1200,
		height: 800,
	}
}

// GenerateBillingReport 生成账单报表图片 - 详细卡片风格（自适应布局）
func (ig *ImageGenerator) GenerateBillingReport(records []models.BillingRecord, month string) (string, error) {
	if len(records) == 0 {
		return "", fmt.Errorf("no records to generate report")
	}

	// 根据记录数量动态调整布局
	recordCount := len(records)
	var cardWidth, cardHeight, cardsPerRow int
	padding := 30
	headerHeight := 120
	summaryHeight := 140 // 新增汇总区域高度

	// 智能布局策略
	if recordCount <= 2 {
		// 1-2条记录：大卡片，横向排列
		cardWidth = 520
		cardHeight = 940
		cardsPerRow = recordCount
	} else if recordCount <= 6 {
		// 3-6条记录：中等卡片，每行2个
		cardWidth = 520
		cardHeight = 940
		cardsPerRow = 2
	} else if recordCount <= 12 {
		// 7-12条记录：小卡片，每行3个
		cardWidth = 420
		cardHeight = 800
		cardsPerRow = 3
	} else {
		// 13+条记录：更小卡片，每行4个
		cardWidth = 350
		cardHeight = 700
		cardsPerRow = 4
	}

	// 计算需要的行数
	rows := (recordCount + cardsPerRow - 1) / cardsPerRow

	// 计算画布总尺寸（包含汇总区域）
	totalWidth := cardsPerRow*cardWidth + (cardsPerRow+1)*padding
	totalHeight := headerHeight + summaryHeight + rows*cardHeight + (rows+1)*padding

	// 限制最大高度，避免图片过大（最大30000像素）
	maxHeight := 30000
	if totalHeight > maxHeight {
		// 如果超过最大高度，减少卡片高度
		scale := float64(maxHeight-headerHeight-summaryHeight-(rows+1)*padding) / float64(rows*cardHeight)
		cardHeight = int(float64(cardHeight) * scale)
		totalHeight = headerHeight + summaryHeight + rows*cardHeight + (rows+1)*padding
	}

	// 创建画布
	dc := gg.NewContext(totalWidth, totalHeight)

	// 绘制渐变背景
	for i := 0; i < totalHeight; i++ {
		ratio := float64(i) / float64(totalHeight)
		r := uint8(95 + ratio*40)
		g := uint8(100 + ratio*60)
		b := uint8(230 - ratio*30)
		dc.SetRGB(float64(r)/255, float64(g)/255, float64(b)/255)
		dc.DrawRectangle(0, float64(i), float64(totalWidth), 1)
		dc.Fill()
	}

	// 绘制标题（字体大小根据宽度调整，增大以便手机查看）
	dc.SetColor(color.White)
	titleSize := 56    // 42 -> 56
	subtitleSize := 24 // 18 -> 24
	if totalWidth < 800 {
		titleSize = 42    // 32 -> 42
		subtitleSize = 18 // 14 -> 18
	} else if totalWidth > 2000 {
		titleSize = 68    // 52 -> 68
		subtitleSize = 28 // 22 -> 28
	}

	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(titleSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(titleSize))
	}
	dc.DrawStringAnchored("水电费账单批量报表", float64(totalWidth)/2, 40, 0.5, 0.5)

	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(subtitleSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(subtitleSize))
	}
	dc.DrawStringAnchored(fmt.Sprintf("账单月份：%s  |  共 %d 条记录  |  布局：%d×%d  |  生成时间：%s",
		month, len(records), cardsPerRow, rows, time.Now().Format("2006-01-02 15:04")),
		float64(totalWidth)/2, 85, 0.5, 0.5)

	// 计算汇总统计
	var totalWaterCost, totalElectricCost, totalManagementFee, totalExtraFee, grandTotal float64
	for _, record := range records {
		totalWaterCost += record.TotalWaterCost
		totalElectricCost += record.TotalElectricCost
		totalManagementFee += record.ManagementFee
		// 计算额外费用
		for _, fee := range record.ExtraFees {
			totalExtraFee += fee.Amount
		}
		grandTotal += record.TotalCost
	}

	// 绘制汇总统计区域
	ig.drawSummarySection(dc, totalWidth, totalWaterCost, totalElectricCost, totalManagementFee, totalExtraFee, grandTotal)

	// 绘制每个详细卡片（在汇总区域下方）
	for i, record := range records {
		row := i / cardsPerRow
		col := i % cardsPerRow

		x := padding + col*(cardWidth+padding)
		y := headerHeight + summaryHeight + padding + row*(cardHeight+padding)

		ig.drawDetailedCard(dc, record, x, y, cardWidth, cardHeight)
	}

	// 保存图片
	filename := fmt.Sprintf("reports/billing_batch_%s_%s.png", month, time.Now().Format("20060102150405"))
	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", err
	}

	if err := dc.SavePNG(filename); err != nil {
		return "", err
	}

	return filename, nil
}

// drawDetailedCard 绘制单个详细卡片（自适应尺寸）
func (ig *ImageGenerator) drawDetailedCard(dc *gg.Context, record models.BillingRecord, x, y, width, height int) {
	fx, fy := float64(x), float64(y)
	fw, fh := float64(width), float64(height)

	// 根据卡片宽度计算缩放因子（基准宽度520）
	scale := float64(width) / 520.0

	// 自适应字体大小（增大以便手机查看）
	titleFontSize := int(36 * scale)        // 24 -> 36
	subtitleFontSize := int(16 * scale)     // 12 -> 16
	headerFontSize := int(28 * scale)       // 20 -> 28
	smallFontSize := int(15 * scale)        // 11 -> 15
	sectionTitleFontSize := int(20 * scale) // 14 -> 20
	infoFontSize := int(17 * scale)         // 12 -> 17

	// 自适应间距和尺寸
	padding := 20.0 * scale
	headerBarHeight := 80.0 * scale
	cornerRadius := 15.0 * scale
	sectionCornerRadius := 8.0 * scale

	// 主卡片 - 白色圆角卡片
	dc.SetColor(color.White)
	dc.DrawRoundedRectangle(fx, fy, fw, fh, cornerRadius)
	dc.Fill()

	// 顶部装饰条 - 渐变色
	headerBarHeightInt := int(headerBarHeight)
	for i := 0; i < headerBarHeightInt; i++ {
		ratio := float64(i) / headerBarHeight
		r := uint8(102 + ratio*50)
		g := uint8(126 + ratio*60)
		b := uint8(234 - ratio*34)
		dc.SetRGB(float64(r)/255, float64(g)/255, float64(b)/255)
		dc.DrawRectangle(fx, fy+float64(i), fw, 1)
		dc.Fill()
	}

	// 标题
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(titleFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(titleFontSize))
	}
	dc.DrawStringAnchored("水电费账单", fx+fw/2, fy+30*scale, 0.5, 0.5)

	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(subtitleFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(subtitleFontSize))
	}
	sendTime := time.Now().Format("2006年01月")
	dc.DrawStringAnchored(fmt.Sprintf("发送时间：%s", sendTime), fx+fw/2, fy+58*scale, 0.5, 0.5)

	// 房间号卡片
	dc.SetColor(color.RGBA{102, 126, 234, 255})
	dc.DrawRoundedRectangle(fx+padding, fy+100*scale, fw-2*padding, 50*scale, sectionCornerRadius)
	dc.Fill()
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(headerFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(headerFontSize))
	}
	dc.DrawStringAnchored(record.RoomNumber, fx+fw/2, fy+112*scale, 0.5, 0.5)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(smallFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(smallFontSize))
	}
	dc.DrawStringAnchored(fmt.Sprintf("账单月份：%s", record.BillingMonth), fx+fw/2, fy+137*scale, 0.5, 0.5)

	// 水费区域
	currentY := fy + 170*scale
	dc.SetColor(color.RGBA{23, 162, 184, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, 28*scale, sectionCornerRadius)
	dc.Fill()
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(sectionTitleFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(sectionTitleFontSize))
	}
	dc.DrawStringAnchored("【水费详情】", fx+fw/2, currentY+14*scale, 0.5, 0.5)

	// 水费信息
	currentY += 35 * scale
	dc.SetColor(color.RGBA{240, 248, 255, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, 160*scale, sectionCornerRadius)
	dc.Fill()

	dc.SetColor(color.RGBA{51, 51, 51, 255})
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
	}

	waterInfo := []struct {
		label string
		value string
	}{
		{"本月水表", fmt.Sprintf("%.0f", record.CurrentWater)},
		{"上月水表", fmt.Sprintf("%.0f", record.PreviousWater)},
		{"水分摊", fmt.Sprintf("%.0f", record.WaterAdjustment)},
		{"水价格", fmt.Sprintf("%.2f 元/吨", record.WaterPrice)},
		{"用水量", fmt.Sprintf("%.1f 吨", record.WaterUsage)},
		{"水费小计", fmt.Sprintf("¥%.2f", record.TotalWaterCost)},
	}

	infoY := currentY + 20*scale
	lineHeight := 25.0 * scale
	for _, info := range waterInfo {
		dc.DrawString(info.label, fx+35*scale, infoY)
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(infoFontSize)); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(infoFontSize))
		}
		dc.DrawStringAnchored(info.value, fx+fw-35*scale, infoY, 1.0, 0.5)
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
		}
		infoY += lineHeight
	}

	// 电费区域
	currentY += 175 * scale
	dc.SetColor(color.RGBA{255, 193, 7, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, 28*scale, sectionCornerRadius)
	dc.Fill()
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(sectionTitleFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(sectionTitleFontSize))
	}
	dc.DrawStringAnchored("【电费详情】", fx+fw/2, currentY+14*scale, 0.5, 0.5)

	// 电费信息
	currentY += 35 * scale
	dc.SetColor(color.RGBA{255, 253, 240, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, 160*scale, sectionCornerRadius)
	dc.Fill()

	dc.SetColor(color.RGBA{51, 51, 51, 255})
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
	}

	electricInfo := []struct {
		label string
		value string
	}{
		{"本月电表", fmt.Sprintf("%.0f", record.CurrentElectric)},
		{"上月电表", fmt.Sprintf("%.0f", record.PreviousElectric)},
		{"电分摊", fmt.Sprintf("%.0f", record.ElectricAdjustment)},
		{"电价格", fmt.Sprintf("%.2f 元/度", record.ElectricPrice)},
		{"用电量", fmt.Sprintf("%.1f 度", record.ElectricUsage)},
		{"电费小计", fmt.Sprintf("¥%.2f", record.TotalElectricCost)},
	}

	infoY = currentY + 20*scale
	for _, info := range electricInfo {
		dc.DrawString(info.label, fx+35*scale, infoY)
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(infoFontSize)); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(infoFontSize))
		}
		dc.DrawStringAnchored(info.value, fx+fw-35*scale, infoY, 1.0, 0.5)
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
		}
		infoY += lineHeight
	}

	// 管理费区域
	currentY += 175 * scale
	dc.SetColor(color.RGBA{108, 117, 125, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, 28*scale, sectionCornerRadius)
	dc.Fill()
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(sectionTitleFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(sectionTitleFontSize))
	}
	dc.DrawStringAnchored("【其他费用】", fx+fw/2, currentY+14*scale, 0.5, 0.5)

	// 管理费和额外费用卡片
	currentY += 35 * scale

	// 计算需要显示的项目数量（管理费 + 额外费用）
	itemCount := 1 + len(record.ExtraFees)
	feeCardHeight := float64(itemCount)*lineHeight + 15*scale

	dc.SetColor(color.RGBA{248, 249, 250, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, feeCardHeight, sectionCornerRadius)
	dc.Fill()
	dc.SetColor(color.RGBA{51, 51, 51, 255})
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
	}

	// 管理费
	itemY := currentY + 20*scale
	dc.DrawString("管理费", fx+35*scale, itemY)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(infoFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(infoFontSize))
	}
	dc.DrawStringAnchored(fmt.Sprintf("¥%.2f", record.ManagementFee), fx+fw-35*scale, itemY, 1.0, 0.5)

	// 额外费用（如果有）
	if len(record.ExtraFees) > 0 {
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
		}
		for _, fee := range record.ExtraFees {
			itemY += lineHeight
			dc.DrawString(fee.Name, fx+35*scale, itemY)
			if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(infoFontSize)); err != nil {
				dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(infoFontSize))
			}
			dc.DrawStringAnchored(fmt.Sprintf("¥%.2f", fee.Amount), fx+fw-35*scale, itemY, 1.0, 0.5)
			if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", float64(infoFontSize)); err != nil {
				dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", float64(infoFontSize))
			}
		}
	}

	// 总费用区域 - 醒目显示
	currentY += feeCardHeight + 10*scale
	totalLabelFontSize := int(22 * scale) // 15 -> 22
	totalValueFontSize := int(30 * scale) // 20 -> 30
	dc.SetColor(color.RGBA{220, 53, 69, 255})
	dc.DrawRoundedRectangle(fx+padding, currentY, fw-2*padding, 42*scale, sectionCornerRadius)
	dc.Fill()
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(totalLabelFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(totalLabelFontSize))
	}
	dc.DrawString("总费用", fx+35*scale, currentY+26*scale)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", float64(totalValueFontSize)); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", float64(totalValueFontSize))
	}
	dc.DrawStringAnchored(fmt.Sprintf("¥%.2f", record.TotalCost), fx+fw-35*scale, currentY+26*scale, 1.0, 0.5)
}

// drawGradientBackground 绘制渐变背景
func (ig *ImageGenerator) drawGradientBackground(dc *gg.Context) {
	// 创建渐变效果
	for i := 0; i < ig.height; i++ {
		ratio := float64(i) / float64(ig.height)
		r := uint8(102 + ratio*16)
		g := uint8(126 + ratio*36)
		b := uint8(234 - ratio*72)
		dc.SetRGB(float64(r)/255, float64(g)/255, float64(b)/255)
		dc.DrawRectangle(0, float64(i), float64(ig.width), 1)
		dc.Fill()
	}

	// 白色内容区域
	dc.SetColor(color.White)
	dc.DrawRoundedRectangle(40, 120, float64(ig.width-80), float64(ig.height-160), 20)
	dc.Fill()
}

// drawTitle 绘制标题
func (ig *ImageGenerator) drawTitle(dc *gg.Context, month string) {
	dc.SetColor(color.White)

	// 加载字体（使用系统默认字体）
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 48); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 48)
	}

	title := "💧⚡ 水电费账单报表"
	dc.DrawStringAnchored(title, float64(ig.width)/2, 70, 0.5, 0.5)
}

// drawStatistics 绘制统计信息
func (ig *ImageGenerator) drawStatistics(dc *gg.Context, count int, totalCost, totalWater, totalElectric float64) {
	startY := 180.0
	cardWidth := 260.0
	cardHeight := 120.0
	gap := 20.0
	startX := 60.0

	stats := []struct {
		label string
		value string
		icon  string
		color color.Color
	}{
		{"记录数", fmt.Sprintf("%d", count), "📊", color.RGBA{102, 126, 234, 255}},
		{"总费用", fmt.Sprintf("¥%.2f", totalCost), "💰", color.RGBA{220, 53, 69, 255}},
		{"总水费", fmt.Sprintf("¥%.2f", totalWater), "💧", color.RGBA{23, 162, 184, 255}},
		{"总电费", fmt.Sprintf("¥%.2f", totalElectric), "⚡", color.RGBA{255, 193, 7, 255}},
	}

	for i, stat := range stats {
		x := startX + float64(i)*(cardWidth+gap)

		// 绘制卡片背景
		dc.SetColor(stat.color)
		dc.DrawRoundedRectangle(x, startY, cardWidth, cardHeight, 15)
		dc.Fill()

		// 绘制图标
		dc.SetColor(color.White)
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\seguiemj.ttf", 36); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 36)
		}
		dc.DrawStringAnchored(stat.icon, x+40, startY+35, 0.5, 0.5)

		// 绘制标签
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 18); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 18)
		}
		dc.DrawStringAnchored(stat.label, x+cardWidth/2, startY+55, 0.5, 0.5)

		// 绘制数值
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 24); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 24)
		}
		dc.DrawStringAnchored(stat.value, x+cardWidth/2, startY+90, 0.5, 0.5)
	}
}

// drawRecords 绘制记录列表
func (ig *ImageGenerator) drawRecords(dc *gg.Context, records []models.BillingRecord) {
	startY := 340.0
	rowHeight := 70.0
	startX := 60.0
	contentWidth := float64(ig.width - 120)

	// 表头
	dc.SetColor(color.RGBA{102, 126, 234, 255})
	dc.DrawRoundedRectangle(startX, startY, contentWidth, 50, 10)
	dc.Fill()

	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 18); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 18)
	}

	headers := []struct {
		text string
		x    float64
	}{
		{"住户", startX + 50},
		{"用水(吨)", startX + 200},
		{"水费(元)", startX + 350},
		{"用电(度)", startX + 500},
		{"电费(元)", startX + 650},
		{"管理费", startX + 800},
		{"总费用", startX + 950},
	}

	for _, h := range headers {
		dc.DrawStringAnchored(h.text, h.x, startY+25, 0.5, 0.5)
	}

	// 数据行
	currentY := startY + 60
	for i, record := range records {
		// 行背景
		if i%2 == 0 {
			dc.SetColor(color.RGBA{248, 249, 250, 255})
		} else {
			dc.SetColor(color.White)
		}
		dc.DrawRoundedRectangle(startX, currentY, contentWidth, rowHeight, 8)
		dc.Fill()

		// 文字颜色
		dc.SetColor(color.RGBA{51, 51, 51, 255})
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 16); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 16)
		}

		data := []struct {
			text string
			x    float64
		}{
			{record.RoomNumber, startX + 50},
			{fmt.Sprintf("%.1f", record.WaterUsage), startX + 200},
			{fmt.Sprintf("%.2f", record.TotalWaterCost), startX + 350},
			{fmt.Sprintf("%.1f", record.ElectricUsage), startX + 500},
			{fmt.Sprintf("%.2f", record.TotalElectricCost), startX + 650},
			{fmt.Sprintf("%.2f", record.ManagementFee), startX + 800},
		}

		for _, d := range data {
			dc.DrawStringAnchored(d.text, d.x, currentY+rowHeight/2, 0.5, 0.5)
		}

		// 总费用加粗
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 18); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 18)
		}
		dc.SetColor(color.RGBA{220, 53, 69, 255})
		dc.DrawStringAnchored(fmt.Sprintf("¥%.2f", record.TotalCost), startX+950, currentY+rowHeight/2, 0.5, 0.5)

		currentY += rowHeight
	}
}

// drawFooter 绘制底部信息
func (ig *ImageGenerator) drawFooter(dc *gg.Context) {
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 14); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 14)
	}

	footer := fmt.Sprintf("生成时间: %s | 水电费计算管理系统", time.Now().Format("2006-01-02 15:04:05"))
	dc.DrawStringAnchored(footer, float64(ig.width)/2, float64(ig.height-30), 0.5, 0.5)
}

// calculateTotals 计算总计
func (ig *ImageGenerator) calculateTotals(records []models.BillingRecord) (totalCost, totalWater, totalElectric float64) {
	for _, r := range records {
		totalCost += r.TotalCost
		totalWater += r.TotalWaterCost
		totalElectric += r.TotalElectricCost
	}
	return
}

// GenerateSimpleCard 生成简单卡片（单个用户）- 简约风格
func (ig *ImageGenerator) GenerateSimpleCard(record models.BillingRecord) (string, error) {
	// 动态高度
	extraCount := len(record.ExtraFees)
	totalHeight := 780 + extraCount*28

	dc := gg.NewContext(480, totalHeight)

	// 浅灰背景
	dc.SetRGB(0.96, 0.97, 0.98)
	dc.Clear()

	// 白色主卡片
	dc.SetRGB(1, 1, 1)
	dc.DrawRoundedRectangle(16, 16, 448, float64(totalHeight-32), 8)
	dc.Fill()

	// 顶部色条
	dc.SetRGB(0.26, 0.6, 0.88) // #4299e1
	dc.DrawRoundedRectangle(16, 16, 448, 70, 8)
	dc.Fill()
	// 底部补方角
	dc.DrawRectangle(16, 56, 448, 30)
	dc.Fill()

	// 标题
	dc.SetRGB(1, 1, 1)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 28); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 28)
	}
	dc.DrawStringAnchored("水电费账单", 240, 42, 0.5, 0.5)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 14); err != nil {
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 14)
	}
	dc.DrawStringAnchored(fmt.Sprintf("%s · %s", record.RoomNumber, record.BillingMonth), 240, 70, 0.5, 0.5)

	y := 110.0
	loadRegular := func(size float64) {
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", size); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", size)
		}
	}
	loadBold := func(size float64) {
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", size); err != nil {
			dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", size)
		}
	}

	// 绘制分区标题
	drawSection := func(title string, r, g, b float64) {
		dc.SetRGB(r, g, b)
		dc.DrawRoundedRectangle(32, y, 416, 28, 4)
		dc.Fill()
		dc.SetRGB(1, 1, 1)
		loadBold(15)
		dc.DrawStringAnchored(title, 240, y+14, 0.5, 0.5)
		y += 36
	}

	// 绘制键值行
	drawRow := func(label, value string) {
		dc.SetRGB(0.29, 0.33, 0.39)
		loadRegular(15)
		dc.DrawStringAnchored(label, 48, y, 0, 0.5)
		loadBold(15)
		dc.DrawStringAnchored(value, 432, y, 1.0, 0.5)
		y += 26
	}

	// 水费
	drawSection("水费明细", 0.09, 0.64, 0.72)
	drawRow("本月水表", fmt.Sprintf("%.0f", record.CurrentWater))
	drawRow("上月水表", fmt.Sprintf("%.0f", record.PreviousWater))
	drawRow("补差", fmt.Sprintf("%.0f", record.WaterAdjustment))
	drawRow("用水量", fmt.Sprintf("%.1f 吨", record.WaterUsage))
	drawRow("水单价", fmt.Sprintf("%.2f 元/吨", record.WaterPrice))
	drawRow("水费小计", fmt.Sprintf("¥%.2f", record.TotalWaterCost))
	y += 10

	// 电费
	drawSection("电费明细", 0.93, 0.55, 0.14)
	drawRow("本月电表", fmt.Sprintf("%.0f", record.CurrentElectric))
	drawRow("上月电表", fmt.Sprintf("%.0f", record.PreviousElectric))
	drawRow("补差", fmt.Sprintf("%.0f", record.ElectricAdjustment))
	drawRow("用电量", fmt.Sprintf("%.1f 度", record.ElectricUsage))
	drawRow("电单价", fmt.Sprintf("%.2f 元/度", record.ElectricPrice))
	drawRow("电费小计", fmt.Sprintf("¥%.2f", record.TotalElectricCost))
	y += 10

	// 其他费用
	drawSection("其他费用", 0.42, 0.46, 0.49)
	drawRow("管理费", fmt.Sprintf("¥%.2f", record.ManagementFee))
	for _, fee := range record.ExtraFees {
		drawRow(fee.Name, fmt.Sprintf("¥%.2f", fee.Amount))
	}
	y += 16

	// 总费用
	dc.SetRGB(0.86, 0.21, 0.27)
	dc.DrawRoundedRectangle(32, y, 416, 44, 6)
	dc.Fill()
	dc.SetRGB(1, 1, 1)
	loadBold(18)
	dc.DrawString("总费用", 52, y+25)
	loadBold(26)
	dc.DrawStringAnchored(fmt.Sprintf("¥%.2f", record.TotalCost), 432, y+24, 1.0, 0.5)

	// 保存
	filename := fmt.Sprintf("reports/card_%s_%s.png", record.RoomNumber, time.Now().Format("20060102150405"))
	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", err
	}
	if err := dc.SavePNG(filename); err != nil {
		return "", err
	}
	return filename, nil
}

// GetImage 读取图片文件
func (ig *ImageGenerator) GetImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// drawSummarySection 绘制汇总统计区域
func (ig *ImageGenerator) drawSummarySection(dc *gg.Context, totalWidth int,
	totalWater, totalElectric, totalManagement, totalExtra, grandTotal float64) {

	summaryY := 115.0 // 汇总区域起始 Y 坐标
	summaryHeight := 120.0
	padding := 30.0

	// 绘制汇总区域背景（渐变绿色）
	for i := 0; i < int(summaryHeight); i++ {
		ratio := float64(i) / summaryHeight
		r := uint8(40 + ratio*20)  // 40-60
		g := uint8(167 - ratio*20) // 167-147
		b := uint8(69 + ratio*82)  // 69-151
		dc.SetRGB(float64(r)/255, float64(g)/255, float64(b)/255)
		dc.DrawRectangle(padding, summaryY+float64(i), float64(totalWidth)-2*padding, 1)
		dc.Fill()
	}

	// 绘制圆角边框
	dc.SetColor(color.RGBA{40, 167, 69, 255})
	dc.DrawRoundedRectangle(padding, summaryY, float64(totalWidth)-2*padding, summaryHeight, 15)
	dc.SetLineWidth(3)
	dc.Stroke()

	// 标题
	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 32); err != nil { // 24 -> 32
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 32)
	}
	dc.DrawStringAnchored("选中记录汇总统计", float64(totalWidth)/2, summaryY+25, 0.5, 0.5)

	// 统计项
	statsY := summaryY + 60
	itemWidth := (float64(totalWidth) - 2*padding - 50) / 6.0 // 6个统计项

	stats := []struct {
		label string
		value string
		icon  string
	}{
		{"总水费", fmt.Sprintf("¥%.2f", totalWater), "💧"},
		{"总电费", fmt.Sprintf("¥%.2f", totalElectric), "⚡"},
		{"管理费", fmt.Sprintf("¥%.2f", totalManagement), "🏢"},
		{"额外费", fmt.Sprintf("¥%.2f", totalExtra), "💵"},
	}

	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 18); err != nil { // 14 -> 18
		dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 18)
	}

	// 绘制前4个统计项
	for i, stat := range stats {
		x := padding + 25 + float64(i)*itemWidth + itemWidth/2

		// 图标
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\seguiemj.ttf", 24); err != nil { // 18 -> 24
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 24)
		}
		dc.DrawStringAnchored(stat.icon, x, statsY-5, 0.5, 0.5)

		// 标签
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyh.ttc", 17); err != nil { // 13 -> 17
			dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", 17)
		}
		dc.DrawStringAnchored(stat.label, x, statsY+18, 0.5, 0.5)

		// 数值
		if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 22); err != nil { // 16 -> 22
			dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 22)
		}
		dc.DrawStringAnchored(stat.value, x, statsY+38, 0.5, 0.5)
	}

	// 合计总费用 - 更大更突出（跨两列显示）
	totalX := padding + 25 + 4*itemWidth + itemWidth

	// 绘制总费用背景高亮
	dc.SetColor(color.RGBA{255, 255, 255, 30})
	dc.DrawRoundedRectangle(totalX-itemWidth*0.9, statsY-18, itemWidth*1.8, 70, 10)
	dc.Fill()

	dc.SetColor(color.White)
	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 24); err != nil { // 18 -> 24
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 24)
	}
	dc.DrawStringAnchored("合计总费用", totalX, statsY+5, 0.5, 0.5)

	if err := dc.LoadFontFace("C:\\Windows\\Fonts\\msyhbd.ttc", 36); err != nil { // 28 -> 36
		dc.LoadFontFace("C:\\Windows\\Fonts\\arialbd.ttf", 36)
	}
	dc.DrawStringAnchored(fmt.Sprintf("¥%.2f", grandTotal), totalX, statsY+35, 0.5, 0.5)
}
