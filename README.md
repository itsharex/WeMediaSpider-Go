# WeMediaSpider-Go

微信公众号文章智能爬虫 Go 版本 - 支持批量爬取、多格式导出、图片下载、数据库存储、专业级安全架构

[![Version](https://img.shields.io/badge/version-1.2.0-blue.svg)](https://github.com/vag-Zhao/WeMediaSpider-Go/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.18+-00ADD8.svg)](https://golang.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red.svg)](https://wails.io/)

## ✨ v1.2.0 重大更新

- 🗄️ **数据库迁移**：从 JSON 文件迁移到 SQLite 数据库，性能大幅提升
- 🔐 **专业级安全**：AES-256-GCM 加密 + HMAC 完整性校验
- 📊 **按公众号分组**：数据按公众号分组显示，更直观
- 🎨 **UI 全面优化**：优化卡片布局、滚动体验、固定标题
- 🛠️ **数据迁移工具**：自动迁移旧版本数据，平滑升级

## 功能特性

### 核心功能
- 🚀 **批量爬取**：支持多个公众号并发爬取，提高效率
- 📊 **多格式导出**：支持 Excel、CSV、JSON、Markdown 等多种格式
- 🖼️ **图片下载**：批量下载文章中的图片，支持进度显示
- 🗄️ **数据库存储**：使用 SQLite 数据库，支持高效查询和搜索
- 💾 **智能缓存**：避免重复请求，提升爬取速度

### 安全特性
- 🔐 **AES-256-GCM 加密**：登录凭证加密存储
- 🔐 **HMAC 完整性校验**：防止文件篡改
- 🔐 **0600 文件权限**：敏感文件仅所有者可访问
- 🔐 **PBKDF2 密钥派生**：100,000 次迭代，安全性更高
- 🔐 **自动格式升级**：旧格式自动升级到新格式

### 数据管理
- 📦 **按公众号分组**：数据按公众号分组显示
- 📦 **时间跨度显示**：显示每个公众号的文章时间范围
- 📦 **容器内滚动**：文章列表支持容器内滚动（400px）
- 📦 **数据迁移工具**：自动迁移 JSON 数据到数据库

### 用户体验
- 🔄 **版本更新**：自动检查新版本并提醒更新
- 🎯 **单实例运行**：防止多个程序同时启动
- ⚙️ **配置持久化**：所有设置自动保存到本地
- 🎨 **现代界面**：基于 React + Ant Design 的现代化桌面应用
- 🪟 **系统托盘**：支持最小化到系统托盘，后台运行

## 技术栈

### 后端
- **Go** - 高性能后端语言
- **Wails v2** - Go 桌面应用框架
- **SQLite + GORM** - 嵌入式数据库 + ORM 框架
- **AES-256-GCM** - 加密算法
- **HMAC-SHA256** - 完整性校验

### 前端
- **React** - UI 框架
- **TypeScript** - 类型安全
- **Ant Design** - UI 组件库
- **Zustand** - 状态管理
- **Vite** - 构建工具

## 快速开始

### 下载使用（推荐）

直接下载编译好的可执行文件：

1. 访问 [Releases 页面](https://github.com/vag-Zhao/WeMediaSpider-Go/releases)
2. 下载最新版本的 `WeMediaSpider-v1.2.0-windows-amd64.tar.gz`
3. 解压后运行 `WeMediaSpider.exe`

### 从 v1.1.0 或更早版本升级

**重要提示：** v1.2.0 包含数据存储格式变更，需要进行数据迁移。

#### 自动迁移（推荐）

1. **备份数据**（可选但推荐）
   ```bash
   cp -r ~/.wemediaspider ~/.wemediaspider.backup
   ```

2. **下载新版本并解压**

3. **运行迁移工具**
   ```bash
   go run backend/cmd/migrate/main.go
   ```

4. **查看迁移报告**
   ```bash
   cat ~/.wemediaspider/backup/migration_report.txt
   ```

5. **启动应用**

#### 全新安装

如果不需要保留旧数据：
```bash
rm -rf ~/.wemediaspider
./WeMediaSpider.exe
```

### 开发环境

#### 环境要求

- Go 1.18+
- Node.js 16+
- Wails CLI

#### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 开发模式

```bash
# 克隆项目
git clone https://github.com/vag-Zhao/WeMediaSpider-Go.git
cd WeMediaSpider-Go

# 运行开发模式
wails dev
```

### 构建应用

```bash
# 构建生产版本
wails build

# 构建后的应用在 build/bin 目录
```

## 数据迁移

### 迁移工具

v1.2.0 提供了自动数据迁移工具，可以将旧版本的 JSON 数据迁移到 SQLite 数据库。

```bash
# 运行迁移工具
go run backend/cmd/migrate/main.go
```

迁移工具会：
1. 自动备份 JSON 文件到 `~/.wemediaspider/backup/`
2. 解析所有 JSON 文件并去重
3. 批量插入数据库（事务保护）
4. 更新统计信息
5. 验证数据完整性
6. 生成迁移报告

### 数据库位置

- **数据库文件**: `~/.wemediaspider/wemedia.db`
- **备份目录**: `~/.wemediaspider/backup/`
- **迁移报告**: `~/.wemediaspider/backup/migration_report.txt`

## 使用说明

1. **登录**：首次使用需要扫码登录微信公众号平台
2. **搜索账号**：输入公众号名称搜索并选择
3. **配置爬取**：设置日期范围、并发数等参数
4. **开始爬取**：点击开始按钮，实时查看进度
5. **查看结果**：在数据界面查看爬取的文章
6. **导出数据**：支持多种格式导出

## 项目结构

```
WeMediaSpider/
├── backend/              # Go 后端代码
│   ├── app/             # 应用主逻辑
│   ├── cmd/             # 命令行工具
│   │   └── migrate/     # 数据迁移工具
│   ├── internal/        # 内部包
│   │   ├── cache/       # 缓存管理
│   │   ├── config/      # 配置管理
│   │   ├── database/    # 数据库模块
│   │   │   └── models/  # 数据库模型
│   │   ├── export/      # 导出功能
│   │   ├── models/      # 数据模型
│   │   ├── repository/  # 数据访问层
│   │   ├── spider/      # 爬虫核心
│   │   └── storage/     # 数据存储（已废弃）
│   └── pkg/             # 公共包
│       ├── crypto/      # 加密工具
│       ├── logger/      # 日志工具
│       └── utils/       # 工具函数
├── frontend/            # React 前端代码
│   ├── src/
│   │   ├── components/  # 组件
│   │   ├── pages/       # 页面
│   │   ├── services/    # API 服务
│   │   ├── stores/      # 状态管理
│   │   └── types/       # 类型定义
│   └── wailsjs/         # Wails 生成的绑定
├── build/               # 构建资源
└── main.go             # 入口文件
```

## 数据存储

应用数据存储在用户主目录的隐藏文件夹中：

- **Windows**: `%USERPROFILE%\.wemediaspider\`
- **macOS**: `~/.wemediaspider/`
- **Linux**: `~/.wemediaspider/`

目录结构：
```
~/.wemediaspider/
├── config.json          # 应用配置（爬取参数、输出目录等）
├── system_config.json   # 系统配置（托盘设置、更新提示等）
├── appdata.json         # 统计数据（文章数、公众号数等）
├── cache.db             # SQLite 缓存数据库
├── login_cache.json     # 登录会话缓存
├── master.key           # AES-256-GCM 加密密钥
└── data/                # 爬取的文章数据目录
    ├── scrape_20260309_120000.json
    ├── scrape_20260309_130000.json
    └── ...
```

**默认导出目录**: `~/Documents/WeMediaSpider/`（可在设置中修改）

## 更新日志

### v1.1.0 (2026-03-09)
- ✨ 新增"今日不再提示"功能，用户可选择暂时忽略更新提醒
- 🎨 重新设计更新弹窗UI，采用简洁的320px宽度设计
- 💾 实现配置持久化到本地文件，所有设置自动保存
- 🎯 实现单实例运行，防止多个程序同时启动
- 🔧 优化启动时更新检查逻辑，尊重用户选择
- 🗑️ 移除更新检查缓存机制，确保实时检测最新版本
- 📝 添加详细的日志记录，便于问题排查

### v1.0.5 (2026-03-08)
- 🪟 完善系统托盘功能和关闭行为
- ⚙️ 优化配置管理系统

### v1.0.4 (2026-03-07)
- 🔄 优化更新检查机制
- 🐛 修复已知问题

## 许可证

MIT License

## 相关项目

- [WeMediaSpider (Python 版本)](https://github.com/vag-Zhao/WeMediaSpider) - 原 Python 版本

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

- Email: zgs3344@hunnu.edu.cn
- GitHub: [@vag-Zhao](https://github.com/vag-Zhao)
