# CloudPulse

一站式的多平台云服务资源监控仪表盘。

## 解决的问题

使用 Vercel、Neon、Supabase、Cloudflare 等免费额度的开发者，需要逐个登录各平台查看用量，效率低且容易超额。CloudPulse 通过 API Key 接入各平台，统一展示用量数据和项目信息，支持额度预警。

## 支持平台

| 平台 | 监控指标 | 项目同步 |
|------|---------|---------|
| Vercel | 带宽、构建次数、Serverless 执行 | 项目列表 + 部署 URL |
| Neon | 存储、计算时间 | 项目列表 + 控制台链接 |
| Supabase | 数据库、带宽、文件存储 | 项目列表 + Dashboard 链接 |
| Cloudflare | R2 存储、D1 数据库、Pages | Pages/R2/D1 资源列表 |

## 技术架构

```
┌─────────────────────────────────────────┐
│  Frontend · React + TypeScript          │
│  Dashboard / ServiceDetail / Settings   │
├──────────── Wails Bindings ─────────────┤
│  Backend · Go                           │
│  ┌─────────┬─────────┬──────────┐       │
│  │  Vercel │  Neon   │Supabase  │       │
│  │  Plugin │  Plugin │  Plugin  │       │
│  └────┬────┴────┬────┴────┬─────┘       │
│       │         │         │             │
│  ConfigManager · CacheManager · PluginManager │
└─────────────────────────────────────────┘
```

- **前端**: React + TypeScript + Vite，支持明暗主题切换
- **后端**: Go + Wails v2，通过 HTTP 调用各平台 REST API
- **插件系统**: 每个平台独立 Plugin，实现统一的 `PlatformPlugin` 接口
- **数据存储**: 凭证和配置存储在本地 `~/.cloudpulse/`，不上传任何数据

## 项目结构

```
cloudpulse/
├── main.go                  # 入口，单实例锁
├── app.go                   # Wails 应用主逻辑
├── core/
│   ├── types.go             # 共享数据类型
│   ├── plugin.go            # PlatformPlugin 接口
│   ├── plugin_manager.go    # 插件注册与管理
│   ├── config.go            # 凭证与设置持久化
│   └── cache.go             # API 响应缓存
├── plugins/
│   ├── vercel/vercel.go     # Vercel 插件
│   ├── neon/neon.go         # Neon 插件
│   ├── supabase/supabase.go # Supabase 插件
│   ├── cloudflare/cloudflare.go # Cloudflare 插件
│   └── custom/              # 自定义服务插件
└── frontend/
    └── src/
        ├── App.tsx           # 主应用
        ├── pages/            # 页面组件
        └── components/       # 通用组件
```

## 部署与运行

### 环境要求

- Go 1.21+
- Node.js 18+
- Wails CLI v2

### 安装依赖

```bash
# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装前端依赖
cd frontend && npm install
```

### 开发模式

```bash
wails dev
```

### 构建生产版本

```bash
wails build
```

构建产物位于 `build/bin/cloudpulse.exe`。

### 运行

直接双击 `cloudpulse.exe`，同一时间只允许运行一个实例。

## 使用方式

1. 启动应用后进入「设置」页面
2. 点击「添加新服务」，选择平台
3. 输入 API Token（各平台获取方式见下表）
4. 点击「添加」完成验证
5. 返回仪表盘查看用量数据

### API Token 获取

| 平台 | 获取地址 |
|------|---------|
| Vercel | https://vercel.com/account/tokens |
| Neon | https://console.neon.tech/app/settings/api-keys |
| Supabase | https://supabase.com/dashboard/account/tokens |
| Cloudflare | https://dash.cloudflare.com/profile/api-tokens |
