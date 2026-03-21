package export

import (
	"fmt"
	"os"
	"strings"

	"WeMediaSpider/backend/internal/models"
	"WeMediaSpider/backend/pkg/logger"

	"go.uber.org/zap"
)

// MarkdownExporter Markdown 导出器
type MarkdownExporter struct{}

// Export 导出为 Markdown
func (e *MarkdownExporter) Export(articles []models.Article, filename string) error {
	logger.Log.Info("📊 开始导出 Markdown 文件", zap.String("file", filename), zap.Int("count", len(articles)))

	file, err := os.Create(filename)
	if err != nil {
		logger.Log.Error("❌ 创建 Markdown 文件失败", zap.Error(err))
		return err
	}
	defer file.Close()

	// 写入标题
	file.WriteString("# 微信公众号文章列表\n\n")

	// 按公众号分组
	accountMap := make(map[string][]models.Article)
	for _, article := range articles {
		accountMap[article.AccountName] = append(accountMap[article.AccountName], article)
	}

	logger.Log.Info("📝 按公众号分组写入", zap.Int("accounts", len(accountMap)))

	// 写入每个公众号的文章
	accountCount := 0
	for accountName, accountArticles := range accountMap {
		accountCount++
		logger.Log.Info("  正在写入公众号", zap.String("account", accountName), zap.Int("no", accountCount), zap.Int("total", len(accountMap)), zap.Int("articles", len(accountArticles)))

		file.WriteString(fmt.Sprintf("## %s\n\n", accountName))
		file.WriteString(fmt.Sprintf("共 %d 篇文章\n\n", len(accountArticles)))

		for i, article := range accountArticles {
			file.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, article.Title))
			file.WriteString(fmt.Sprintf("- **发布时间**: %s\n", article.PublishTime))
			file.WriteString(fmt.Sprintf("- **链接**: [点击查看](%s)\n", article.Link))

			// 如果有正文内容
			if article.Content != "" {
				file.WriteString("\n**正文内容**:\n\n")
				file.WriteString(article.Content)
				file.WriteString("\n")
			}

			file.WriteString("\n---\n\n")
		}
	}

	logger.Log.Info("✅ Markdown 文件导出成功", zap.String("file", filename))
	return nil
}

// ExportSingle 导出单篇文章为 Markdown
func (e *MarkdownExporter) ExportSingle(article models.Article, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入标题
	file.WriteString(fmt.Sprintf("# %s\n\n", article.Title))

	// 写入元信息
	file.WriteString("## 文章信息\n\n")
	file.WriteString(fmt.Sprintf("- **公众号**: %s\n", article.AccountName))
	file.WriteString(fmt.Sprintf("- **发布时间**: %s\n", article.PublishTime))
	file.WriteString(fmt.Sprintf("- **原文链接**: [点击查看](%s)\n", article.Link))

	// 写入摘要
	if article.Digest != "" {
		file.WriteString("\n## 摘要\n\n")
		file.WriteString(article.Digest)
		file.WriteString("\n")
	}

	// 写入正文
	if article.Content != "" {
		file.WriteString("\n## 正文\n\n")
		file.WriteString(article.Content)
		file.WriteString("\n")
	}

	return nil
}

// ExportByAccount 按公众号分别导出
func (e *MarkdownExporter) ExportByAccount(articles []models.Article, outputDir string) error {
	// 按公众号分组
	accountMap := make(map[string][]models.Article)
	for _, article := range articles {
		accountMap[article.AccountName] = append(accountMap[article.AccountName], article)
	}

	// 为每个公众号创建文件
	for accountName, accountArticles := range accountMap {
		// 清理文件名
		safeName := strings.ReplaceAll(accountName, "/", "_")
		safeName = strings.ReplaceAll(safeName, "\\", "_")
		safeName = strings.ReplaceAll(safeName, ":", "_")

		filename := fmt.Sprintf("%s/%s.md", outputDir, safeName)
		if err := e.Export(accountArticles, filename); err != nil {
			return err
		}
	}

	return nil
}
