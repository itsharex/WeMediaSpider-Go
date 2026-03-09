# WeMediaSpider-Go

微信公众号文章智能爬虫 Go 版本 - 支持批量爬取、多格式导出、图片下载、智能缓存、桌面应用

[![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)](https://github.com/vag-Zhao/WeMediaSpider-Go/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.18+-00ADD8.svg)](https://golang.org/)
[![Wails](https://img.shields.io/badge/Wails-v2-red.svg)](https://wails.io/)

## 功能特性

- 🚀 **批量爬取**：支持多个公众号并发爬取，提高效率
- 📊 **多格式导出**：支持 Excel、CSV、JSON、Markdown 等多种格式
- 🖼️ **图片下载**：批量下载文章中的图片，支持进度显示
- 🔐 **安全加密**：登录凭证采用 AES-256-GCM 加密存储
- 💾 **智能缓存**：避免重复请求，提升爬取速度
- 📦 **数据管理**：自动保存数据，支持覆盖/追加导入模式
- 🔄 **版本更新**：自动检查新版本并提醒更新，支持"今日不再提示"
- 🎯 **单实例运行**：防止多个程序同时启动，避免数据冲突
- ⚙️ **配置持久化**：所有设置自动保存到本地，重启后保持
- 🎨 **现代界面**：基于 React + Ant Design 的现代化桌面应用
- 🪟 **系统托盘**：支持最小化到系统托盘，后台运行

## 技术栈

### 后端
- **Go** - 高性能后端语言
- **Wails v2** - Go 桌面应用框架
- **AES-256-GCM** - 加密算法

### 前端
- **React** - UI 框架
- **TypeScript** - 类型安全
- **Ant Design** - UI 组件库
- **Zustand** - 状态管理
- **Vite** - 构建工具

## 快速开始

### 环境要求

- Go 1.18+
- Node.js 16+
- Wails CLI

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 开发模式

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
│   ├── internal/        # 内部包
│   │   ├── cache/       # 缓存管理
│   │   ├── config/      # 配置管理
│   │   ├── export/      # 导出功能
│   │   ├── models/      # 数据模型
│   │   ├── spider/      # 爬虫核心
│   │   └── storage/     # 数据存储
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
