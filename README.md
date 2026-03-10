# AirGate SDK

AirGate 插件开发 SDK，定义插件与 Core 之间的全部接口契约。

## 项目概览

AirGate 是一个可扩展的 AI 网关平台，由以下仓库组成：

| 仓库 | 职责 |
| --- | --- |
| `airgate-core` | 运行时引擎：管理后台、账号调度、计费、插件生命周期管理 |
| **`airgate-sdk`** | **接口契约：插件接口、共享类型、gRPC 协议定义** |
| `airgate-openai` | 参考实现：OpenAI 兼容网关插件 |

> SDK 不包含任何业务逻辑，只定义协议边界。

## 安装

```bash
go get github.com/DouDOU-start/airgate-sdk@latest
```

## 架构

```text
┌─────────────────────────────────────────────────┐
│                  airgate-core                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ 账号调度  │ │ 计费限流  │ │  路由 & API 网关  │ │
│  └────┬─────┘ └────┬─────┘ └───────┬──────────┘ │
│       └────────────┼───────────────┘             │
│                    │ gRPC (go-plugin)             │
└────────────────────┼─────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
   ┌─────────┐               ┌─────────┐
   │ 网关插件 │               │ 扩展插件 │
   └─────────┘               └─────────┘
   AI API 代理                支付/监控/管理等
```

Core 通过 [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) 以独立进程方式运行插件，插件与 Core 之间通过 gRPC 通信。

## 核心概念

### 职责分工

```text
┌──────────── Core ────────────┐     ┌────────── 网关插件 ──────────┐
│                               │     │                               │
│  账号管理（增删改查、存储）     │     │  声明账号格式（AccountTypes）  │
│  账号调度（负载均衡、选号）     │     │  声明支持的模型（Models）      │
│  路由注册（HTTP 网关、鉴权）   │     │  声明 API 端点（Routes）       │
│  计费、限流、并发控制          │     │  转发请求到上游（Forward）      │
│                               │     │  返回 token 用量（Result）     │
└───────────────────────────────┘     └───────────────────────────────┘
         通用平台能力                        上游 API 适配器
```

网关插件是一个**适配器**：告诉 Core 这个平台有什么模型、什么端点、账号怎么填，然后负责把请求翻译成上游 API 的格式并转发。Core 做所有通用逻辑。

### 请求生命周期

```text
用户请求 → Core 鉴权 → Core 选账号 → 插件 Forward() → 上游 AI API
                                          ↓
                        Core 记录用量 ← ForwardResult（token 数）
```

### 账号数据结构

Core 使用**一张 accounts 表**存储所有平台的账号，通过 `platform` + `type` 区分：

```text
accounts 表（Core 数据库）
┌────────────────────────────────────────────────────────────┐
│ 基础标识    id, name, platform, type                        │
│ 凭证       credentials (JSONB，结构由 type 决定)             │
│ 网络       proxy_url                                       │
│ 调度参数    status, priority, rate_multiplier, max_concurrency│
│ 运行时状态  expires_at, rate_limited_at, error_message, ...  │
│            （调度参数 + 运行时状态均为 Core 内部管理）          │
└────────────────────────────────────────────────────────────┘
```

**SDK Account 是 Core 传给插件的"最小视图"**，只包含插件转发时需要的字段：

```go
type Account struct {
    ID          int64             `json:"id"`
    Name        string            `json:"name"`
    Platform    string            `json:"platform"`
    Type        string            `json:"type"`        // 账号类型（对应 AccountType.Key，如 "apikey"、"oauth"）
    Credentials map[string]string `json:"credentials"` // JSONB 透传，结构由 Type 决定
    ProxyURL    string            `json:"proxy_url"`
}
```

> 插件拿到的账号已经是 Core 调度好的结果。调度参数（`status`、`priority`、`rate_multiplier`、`max_concurrency`）是 Core 选号的输入，插件不需要感知；运行时状态（`expires_at`、`rate_limited_at`）是 Core 内部管理的，前端直接查库渲染。

**设计原则**：

- **SDK Account 最小化** — 插件只需"用什么凭证调上游"，调度和计费参数全部留在 Core
- **Core 表可自由扩展** — Core 的 accounts 表字段远多于 SDK Account，前端页面直接查库渲染完整数据
- **`type` 驱动凭证解析** — 插件通过 `AccountTypes` 声明凭证模式，Core 存储时带上 `type`，插件 Forward 时据此解析 `credentials`
- **插件 Widget 按需注入** — 运行时字段（过期时间、限流状态等）需要自定义展示时，Core 作为 props 传给插件的 Widget 组件

### 关键标识

| 标识 | 含义 | 示例 |
| --- | --- | --- |
| `PluginInfo.ID` | 运行时唯一标识，Core 用于 API 路径、资源挂载、缓存键 | `"gateway-openai"` |
| `Platform()` | 业务平台键，用于账号关联、调度、计费 | `"openai"` |

> `ID` 和 `Platform` 是不同维度：一个插件 ID 对应一个进程，一个 Platform 对应一个上游平台。

### plugin.yaml

`plugin.yaml` 是由插件代码生成的**分发文件**，仅用于安装和市场展示。**运行时真相始终在插件代码里**，Core 不依赖 `plugin.yaml` 做运行时决策。

## 插件类型

SDK 定义两大类插件：

| 类型 | 接口 | 定位 |
| --- | --- | --- |
| 网关 | `GatewayPlugin` | AI API 代理。插件声明模型和路由、实现转发，Core 自动处理调度/计费/限流 |
| 扩展 | `ExtensionPlugin` | 一切非网关功能。拥有路由注册、数据库迁移、后台任务三大能力 |

> 只定义有实际需求的类型。网关专注 AI API 代理，其余全部归入扩展。

### 网关插件

`GatewayPlugin` 需要实现的核心方法：

| 方法 | 职责 |
| --- | --- |
| `Platform()` | 返回业务平台标识（如 `"openai"`） |
| `Models()` | 声明支持的模型列表（含价格，Core 用于计费） |
| `Routes()` | 声明 API 端点（如 `POST /v1/chat/completions`），Core 自动注册路由 |
| `Forward(ctx, req)` | 拿到 Core 调度好的账号，转发请求到上游，返回 token 用量 |
| `ValidateAccount(ctx, credentials)` | 验证凭证有效性，Core 在添加/导入账号时调用 |
| `HandleWebSocket(ctx, conn)` | 处理 WebSocket 双向通信（如 Realtime API） |

Core 自动处理的能力：

- **账号调度** — 根据负载和可用性选择上游账号
- **计费** — 基于 `ForwardResult` 中的 token 数自动计费
- **限流** — 按用户/分组维度限流
- **并发控制** — 按账号维度控制并发

### 扩展插件

`ExtensionPlugin` 提供三大基础能力，组合使用可覆盖各种非网关场景：

| 能力 | 方法 | 说明 |
| --- | --- | --- |
| 自定义路由 | `RegisterRoutes(r)` | 注册任意 HTTP API |
| 数据库迁移 | `Migrate(db)` | 创建插件专属数据表 |
| 后台任务 | `BackgroundTasks()` | 声明定时任务，Core 负责调度 |

典型场景：

- **支付接入** — 注册支付路由（创建订单、回调）+ 订单表 + 定时对账任务
- **监控告警** — 注册告警配置 API + 告警规则表 + 定时检查任务
- **数据分析** — 注册报表 API + 聚合表 + 定时聚合任务

### 可选能力

所有插件类型都可以额外实现以下可选接口，Core 通过类型断言自动检测：

| 接口 | 用途 | 典型场景 |
| --- | --- | --- |
| `WebAssetsProvider` | 提供前端静态资源 | 自定义管理页面、嵌入组件 |
| `ConfigWatcher` | 配置热更新 | 不重启改配置 |

> OAuth 授权是"添加账号"的一种方式，在插件的账号表单 Widget 中实现，不需要独立接口。

## 快速开始：网关插件

```go
package main

import (
    "context"
    sdk "github.com/DouDOU-start/airgate-sdk"
    "github.com/DouDOU-start/airgate-sdk/grpc"
)

type MyGateway struct{}

// ---- Plugin 基础接口 ----

func (g *MyGateway) Info() sdk.PluginInfo {
    return sdk.PluginInfo{
        ID:      "gateway-myplatform",
        Name:    "My Platform 网关",
        Version: "1.0.0",
        Type:    sdk.PluginTypeGateway,
        AccountTypes: []sdk.AccountType{
            {
                Key:   "apikey",
                Label: "API Key",
                Fields: []sdk.CredentialField{
                    {Key: "api_key", Label: "API Key", Type: "password", Required: true},
                    {Key: "base_url", Label: "API 地址", Type: "text"},
                },
            },
        },
    }
}

func (g *MyGateway) Init(ctx sdk.PluginContext) error { return nil }
func (g *MyGateway) Start(_ context.Context) error   { return nil }
func (g *MyGateway) Stop(_ context.Context) error     { return nil }

// ---- GatewayPlugin 接口 ----

func (g *MyGateway) Platform() string { return "myplatform" }

func (g *MyGateway) Models() []sdk.ModelInfo {
    return []sdk.ModelInfo{
        {ID: "my-model-v1", Name: "My Model V1", MaxTokens: 128000, InputPrice: 1.0, OutputPrice: 3.0},
    }
}

func (g *MyGateway) Routes() []sdk.RouteDefinition {
    return []sdk.RouteDefinition{
        {Method: "POST", Path: "/v1/chat/completions", Description: "Chat API"},
    }
}

func (g *MyGateway) Forward(ctx context.Context, req *sdk.ForwardRequest) (*sdk.ForwardResult, error) {
    // req.Account — Core 已调度好的上游账号
    // req.Body / req.Headers — 原始请求
    // req.Writer — 流式写入 SSE 响应
    return &sdk.ForwardResult{
        StatusCode:   200,
        InputTokens:  100,
        OutputTokens: 50,
        Model:        "my-model-v1",
    }, nil
}

func (g *MyGateway) ValidateAccount(ctx context.Context, credentials map[string]string) error {
    // 用凭证调一次上游接口，验证是否有效
    return nil
}

func (g *MyGateway) HandleWebSocket(ctx context.Context, conn sdk.WebSocketConn) error {
    // 处理 WebSocket 双向通信（不需要可返回 ErrNotSupported）
    return sdk.ErrNotSupported
}

// ---- 启动 ----

func main() {
    grpc.Serve(&MyGateway{})  // 自动识别插件类型，注册 gRPC 服务
}
```

构建：

```bash
go build -o my-plugin .
```

## 前端集成

插件的前端能力分两种：**独立页面**和**组件嵌入**，两者通过同一套资源机制（`WebAssetsProvider`）提供 JS/CSS 文件。

| 模式 | 说明 | 谁控制布局 |
| --- | --- | --- |
| 独立页面 (`FrontendPages`) | 插件拥有完整页面，Core 分配路由和导航入口 | 插件 |
| 组件嵌入 (`FrontendWidgets`) | 插件提供组件片段，嵌入 Core 已有页面的指定插槽 | Core |

### 独立页面

在 `PluginInfo` 中声明，Core 据此注册前端路由并渲染侧边栏导航：

```go
FrontendPages: []sdk.FrontendPage{
    {Path: "/dashboard", Title: "仪表盘", Icon: "chart", Description: "使用统计"},
    {Path: "/settings", Title: "高级设置", Icon: "settings", Description: "插件配置"},
},
```

### 组件嵌入

Core 页面预留了**插槽（Slot）**，插件声明往哪个插槽注入自定义组件。SDK 用常量定义所有合法插槽：

```go
const (
    SlotAccountForm   = "account-form"    // 添加/编辑账号表单
    SlotAccountDetail = "account-detail"  // 账号详情/用量展示
)
```

在 `PluginInfo` 中声明要嵌入的组件：

```go
FrontendWidgets: []sdk.FrontendWidget{
    {Slot: sdk.SlotAccountForm, EntryFile: "widgets/account-form.js", Title: "账号添加"},
    {Slot: sdk.SlotAccountDetail, EntryFile: "widgets/account-detail.js", Title: "用量详情"},
},
```

**与声明式字段的关系**：简单插件只用 `AccountTypes` + `CredentialFields`，Core 即可自动渲染表单。当插件需要超出简单表单的自定义 UI（OAuth 授权按钮、用量图表等）时，用 Widget 覆盖默认渲染：

```text
渲染账号管理页
  → 插件声明了对应 Slot 的 Widget？
    → 是 → 动态加载 EntryFile 指向的 JS 组件
    → 否 → 用 CredentialFields 声明式默认渲染
```

### 静态资源

无论独立页面还是组件嵌入，前端资源都通过 `WebAssetsProvider` 统一提供：

```go
import "embed"

//go:embed web/dist/*
var webAssets embed.FS

func (g *MyGateway) GetWebAssets() map[string][]byte {
    files := make(map[string][]byte)
    // 遍历 embed.FS 构建文件映射
    // files["index.js"] = ...
    // files["widgets/account-form.js"] = ...
    return files
}
```

### 构建与加载流程

```text
构建阶段：
  web/src → npm run build → web/dist → go:embed → 插件二进制

运行阶段：
  Core 启动插件
    → Info() 获取 FrontendPages + FrontendWidgets
    → GetWebAssets() 提取静态资源到本地
    → 用户访问时按需加载
```

## 目录结构

### SDK 仓库

```text
airgate-sdk/
├── plugin.go          # Plugin 基础接口、PluginInfo、可选接口
├── gateway.go         # GatewayPlugin 接口
├── models.go          # 共享类型：Account, ModelInfo, ForwardRequest/Result 等
├── config.go          # PluginConfig 配置读取接口
├── extension.go       # ExtensionPlugin 接口
├── oauth.go           # OAuthHandler 接口
├── websocket.go       # WebSocketHandler / WebSocketConn
├── grpc/              # gRPC 桥接层（go-plugin 适配）
│   ├── go_plugin.go   # Serve() 入口
│   ├── common.go      # pluginBase 公共基类
│   └── *_client.go    # 各插件类型的 gRPC 客户端/服务端
├── shared/            # 握手配置
├── proto/             # protobuf 定义
└── docs/              # 详细文档
```

### 推荐的插件项目结构

```text
my-plugin/
├── cmd/server/main.go        # 入口
├── internal/
│   └── gateway/
│       ├── gateway.go        # 接口实现
│       ├── metadata.go       # 元信息（推荐单源维护）
│       └── assets.go         # WebAssetsProvider + go:embed（可选）
├── web/                       # 前端源码（可选）
│   ├── src/
│   │   ├── pages/            # 独立页面
│   │   └── widgets/          # 嵌入组件
│   ├── package.json
│   └── dist/                  # 构建产物
├── go.mod
└── plugin.yaml                # 由代码生成的分发文件
```

## 打包与发布

### plugin.yaml 示例

```yaml
id: gateway-myplatform
name: My Platform 网关
version: 1.0.0
type: gateway
min_core_version: "1.0.0"
platform: myplatform
routes:
  - method: POST
    path: /v1/chat/completions
models:
  - id: my-model-v1
    name: My Model V1
    max_tokens: 128000
    input_price: 1.0
    output_price: 3.0
account_types:
  - key: apikey
    label: API Key
    fields:
      - key: api_key
        label: API Key
        type: password
        required: true
```

### 打包格式

```text
my-plugin.tar.gz
├── my-plugin           # 插件二进制
├── plugin.yaml         # 分发元信息
└── web/                # 可选：前端资源
```

### 发布检查清单

- [ ] `go test ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] 重新生成最新 `plugin.yaml`
- [ ] 构建目标平台二进制
- [ ] 如有前端，构建前端资源
- [ ] 打包并验证完整性

## 开发工具

```bash
make lint    # 代码检查
make fmt     # 代码格式化
make test    # 运行测试
make proto   # 重新生成 protobuf 代码
```

## 给 Core 开发者

Core 启动插件后的消费流程：

```text
启动插件进程（go-plugin）
  → Info()      获取元信息（ID、类型、账号格式、前端声明）
  → Platform()  获取业务平台键
  → Models()    获取模型列表（缓存，用于计费）
  → Routes()    获取路由声明（注册到 HTTP 网关）
  → GetWebAssets()  提取前端资源（如有）

请求到达时：
  → Core 鉴权、限流
  → Core 调度账号
  → Forward(ctx, req)  调用插件转发
  → Core 记录用量（基于 ForwardResult）
```

Core 必须遵守的约定：

- 以 `PluginInfo.ID` 作为运行时键（API 路径、资源挂载、缓存）
- 以 `Platform()` 作为业务键（账号关联、调度、计费）
- 以插件运行时返回的元信息为准，**不依赖 `plugin.yaml` 做运行时决策**
- 渲染插槽时，优先加载插件的 `FrontendWidgets`，无 Widget 则用 `CredentialFields` 默认渲染

## 文档

| 文档 | 内容 |
| --- | --- |
| [docs/interfaces.md](docs/interfaces.md) | 全部接口定义和类型速览 |
| [docs/plugin-spec.md](docs/plugin-spec.md) | 插件开发规范和发布流程 |
| [docs/grpc-protocol.md](docs/grpc-protocol.md) | gRPC 协议定义 |

## License

MIT
