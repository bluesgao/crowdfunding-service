# 众筹后端服务 (Crowdfunding Service)

基于 Gin + Viper + GORM + go-ethereum 构建的区块链众筹后端服务。

## 功能特性

### 核心功能

- **项目管理**: 创建、更新、查询众筹项目
- **链上事件监控**: 实时监控区块链事件并写入数据库
- **定时任务**: 使用 [gocron](https://github.com/go-co-op/gocron) 管理项目状态
- **贡献记录**: 跟踪和管理项目贡献

### 技术栈

- **Web框架**: Gin
- **配置管理**: Viper
- **数据库ORM**: GORM
- **区块链交互**: go-ethereum
- **定时任务**: gocron v2
- **数据库**: PostgreSQL

## 项目结构

```
crowdfunding-service/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── database/
│   │   └── database.go          # 数据库连接
│   ├── models/
│   │   └── project.go           # 数据模型
│   ├── ethereum/
│   │   ├── client.go            # 以太坊客户端
│   │   └── monitor.go           # 事件监控
│   ├── handlers/
│   │   ├── project.go           # 项目API处理器
│   │   └── blockchain.go        # 区块链API处理器
│   ├── router/
│   │   └── router.go            # 路由配置
│   └── task/
│       └── manager.go           # 任务管理器
├── config/
│   └── config.yaml              # 配置文件
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

### 1. 环境要求

- Go 1.24.5+
- PostgreSQL 12+
- 以太坊节点或Infura等RPC服务

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置数据库

创建PostgreSQL数据库：

```sql
CREATE DATABASE crowdfunding;
```

### 4. 配置文件

复制配置文件模板：

```bash
cp config/config.example.yaml config/config.yaml
```

编辑 `config/config.yaml` 并填入真实配置：

```yaml
server:
  port: "8080"
  mode: "debug"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_password"
  dbname: "crowdfunding"
  sslmode: "disable"

ethereum:
  rpc_url: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
  private_key: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
  contract_address: "0x1234567890123456789012345678901234567890"
  start_block: 0
  confirmations: 12

task:
  interval: 60
```

**重要配置说明：**

- `private_key`: 必须是64位十六进制字符串（32字节），可以带或不带0x前缀
- `contract_address`: 智能合约的部署地址
- `rpc_url`: 以太坊节点RPC地址，可以使用Infura、Alchemy等服务
- **安全提醒**: 私钥包含敏感信息，请确保不要提交到版本控制系统

### 日志配置

服务使用自定义日志器，支持以下配置：

```yaml
log:
  level: "info"        # 日志级别: debug, info, warn, error, fatal
  output: "stdout"     # 输出目标: stdout, stderr, file
  file: "logs/app.log" # 日志文件路径（当output为file时使用）
```

**日志格式**: `[时间] [级别] [文件:行号] 消息`

**示例输出**:
```
[2025-09-12 09:46:02] [INFO] [main.go:47] Server starting on port 8080
[2025-09-12 09:46:02] [ERROR] [database.go:25] Failed to connect to database
```

### 5. 运行服务

```bash
go run cmd/server/main.go
```

## API 接口

### 项目管理

#### 创建项目

```http
POST /api/v1/projects
Content-Type: application/json

{
  "title": "项目标题",
  "description": "项目描述",
  "target_amount": 100.0,
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-12-31T23:59:59Z",
  "creator_address": "0x...",
  "category": "科技"
}
```

#### 获取项目列表

```http
GET /api/v1/projects?page=1&page_size=10&status=active&category=科技
```

#### 获取项目详情

```http
GET /api/v1/projects/{id}
```

#### 更新项目

```http
PUT /api/v1/projects/{id}
Content-Type: application/json

{
  "title": "新标题",
  "description": "新描述"
}
```

#### 取消项目

```http
DELETE /api/v1/projects/{id}
```

#### 获取项目贡献记录

```http
GET /api/v1/projects/{id}/contributions?page=1&page_size=20
```

#### 获取项目统计

```http
GET /api/v1/projects/{id}/stats
```

### 区块链相关

#### 获取区块链状态

```http
GET /api/v1/blockchain/status
```

#### 获取事件列表

```http
GET /api/v1/blockchain/events?type=ContributionMade&project_id=1&processed=true
```

#### 获取事件详情

```http
GET /api/v1/blockchain/events/{id}
```

## 数据模型

### Project (项目)

- `id`: 项目ID
- `title`: 项目标题
- `description`: 项目描述
- `target_amount`: 目标金额
- `current_amount`: 当前金额
- `start_time`: 开始时间
- `end_time`: 结束时间
- `status`: 项目状态 (pending/active/success/failed/cancelled)
- `creator_address`: 创建者地址

### Contribution (贡献)

- `id`: 贡献ID
- `project_id`: 项目ID
- `amount`: 贡献金额
- `address`: 贡献者地址
- `tx_hash`: 交易哈希
- `block_num`: 区块号

### Event (事件)

- `id`: 事件ID
- `project_id`: 项目ID
- `event_type`: 事件类型
- `tx_hash`: 交易哈希
- `block_num`: 区块号
- `data`: 事件数据
- `processed`: 是否已处理

## 定时任务

服务使用 gocron 实现以下定时任务：

1. **项目状态更新** (每60秒)

   - 检查项目开始时间，更新状态为 active
   - 检查项目结束时间，更新状态为 success/failed
   - 检查目标金额达成情况
2. **事件处理** (每30秒)

   - 处理未确认的区块链事件
   - 重新处理失败的事件
3. **数据清理** (每天凌晨2点)

   - 清理30天前的已处理事件
   - 清理已取消/失败项目的旧贡献记录

## 区块链事件

服务监控以下智能合约事件：

- `ContributionMade`: 贡献事件
- `ProjectStatusChanged`: 项目状态变更
- `ProjectCreated`: 项目创建

## 开发说明

### 添加新的API接口

1. 在 `internal/handlers/` 中添加处理器
2. 在 `internal/router/router.go` 中注册路由
3. 更新API文档

### 添加新的定时任务

1. 在 `internal/task/manager.go` 中添加任务函数
2. 在 `Start()` 方法中注册任务

### 添加新的区块链事件

1. 在 `internal/ethereum/client.go` 中添加事件解析
2. 在 `internal/ethereum/monitor.go` 中添加事件处理逻辑

## 部署

### Docker 部署

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
CMD ["./main"]
```

### 环境变量

可以通过环境变量覆盖配置：

```bash
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export SERVER_PORT=8080
export ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROJECT_ID
```

## 许可证

MIT License
