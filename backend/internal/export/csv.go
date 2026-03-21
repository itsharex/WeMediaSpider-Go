package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

// CSVExporter CSV 导出器
type CSVExporter struct{}

// Export 导出为 CSV
func (e *CSVExporter) Export(articles []models.Article, filename string) error {
	logger.Log.Info("📊 开始导出 CSV 文件", zap.String("file", filename), zap.Int("count", len(articles)))

	file, err := os.Create(filename)
	if err != nil {
		logger.Log.Error("❌ 创建 CSV 文件失败", zap.Error(err))
		return err
	}
	defer file.Close()

	// 写入 UTF-8 BOM（Excel 兼容）
	file.Write([]byte{0xEF, 0xBB, 0xBF})
	logger.Log.Info("✅ 已写入 UTF-8 BOM 标记（Excel 兼容）")

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{
		"公众号名称",
		"文章标题",
		"文章链接",
		"发布时间",
		"正文内容",
	}
	if err := writer.Write(headers); err != nil {
		logger.Log.Error("❌ 写入 CSV 表头失败", zap.Error(err))
		return err
	}

	logger.Log.Info("📝 开始写入文章数据", zap.Int("count", len(articles)))

	// 写入数据
	for i, article := range articles {
		record := []string{
			article.AccountName,
			article.Title,
			article.Link,
			article.PublishTime,
			article.Content,
		}
		if err := writer.Write(record); err != nil {
			logger.Log.Error("❌ 写入文章失败", zap.Int("index", i+1), zap.Error(err))
			return err
		}

		if (i+1)%10 == 0 || i == len(articles)-1 {
			logger.Log.Info(fmt.Sprintf("  已写入 %d/%d 篇文章", i+1, len(articles)))
		}
	}

	logger.Log.Info("✅ CSV 文件导出成功", zap.String("file", filename))
	return nil
}
