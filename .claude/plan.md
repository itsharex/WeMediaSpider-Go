# 定时任务页面重构计划

## 问题
当前 SchedulePage 要求用户手动输入 cron 表达式（如 `0 0 2 * * *`），普通用户无法理解和使用。爬取配置也是 JSON 文本框，完全不可用。

## 设计方案
将创建/编辑任务的 Modal 改为直观的表单，用下拉选择器替代 cron 表达式，用与 ScrapePage 一致的表单控件替代 JSON 配置。

### 前端改动

**SchedulePage.tsx — 重写创建/编辑 Modal 表单：**

1. 频率选择（替代 cron 表达式）：
   - `频率` Select：每天 / 每周 / 每隔N小时
   - `执行时间` TimePicker：选择时/分（每天/每周模式）
   - `星期` Select：周一~周日（仅每周模式显示）
   - `间隔小时` InputNumber：2-24（仅每隔N小时模式显示）
   - 前端根据这些字段自动生成 cron 表达式传给后端，用户完全不接触 cron

2. 爬取配置（替代 JSON 文本框）：
   - `公众号列表` TextArea：每行一个，与 ScrapePage 一致
   - `采集天数` InputNumber：采集最近N天的文章（默认30），替代日期范围
   - `最大页数` InputNumber（默认10）
   - `请求间隔` InputNumber 秒（默认10）
   - `获取正文` Switch

3. 任务卡片优化：
   - 将 cron 表达式显示改为中文描述（如"每天 02:00"、"每周一 09:30"）

### 后端改动
无需改动。后端已经完整支持 cron 表达式和 JSON scrapeConfig，前端只需正确生成这两个字段即可。

### 文件清单
- `frontend/src/pages/SchedulePage.tsx` — 重写（唯一需要改的文件）
