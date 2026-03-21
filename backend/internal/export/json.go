package export

import (
	"encoding/json"
	"os"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

// JSONExporter JSON 导出器
type JSONExporter struct{}

// Export 导出为 JSON
func (e *JSONExporter) Export(articles []models.Article, filename string) error {
	logger.Log.Info("开始导出 JSON 文件", zap.String("file", filename), zap.Int("count", len(articles)))

	file, err := os.Create(filename)
	if err != nil {
		logger.Log.Error("创建 JSON 文件失败", zap.String("file", filename), zap.Error(err))
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(articles); err != nil {
		logger.Log.Error("编码 JSON 数据失败", zap.Error(err))
		return err
	}

	logger.Log.Info("JSON 文件导出成功", zap.String("file", filename))
	return nil
}
