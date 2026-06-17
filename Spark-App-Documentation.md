# Spark App - 项目文档

> Soul-like Gen Z 社交交友 App | Flutter + Go | 跨平台 iOS/Android

---

## 1. 项目概述

**Spark** 是一款面向 Z 世代年轻人的社交交友应用，对标 Soul。核心功能包括虚拟形象、匿名社交、兴趣匹配和一对一实时聊天。

| 维度 | 选型 |
|---|---|
| 前端 | Flutter 3.x + Riverpod + go_router |
| 后端 | Go 1.22+ + Gin + GORM |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 消息队列 | Redis List（MVP）；日活破 5000 后引入分布式消息队列 |
| 推送 | JPush |
| 存储 | Aliyun OSS |
| WebSocket | gorilla/websocket |

---

## 2. 项目结构

```
spark-app/
├── backend/                          # Go 后端
│   ├── cmd/server/main.go            # 入口
│   ├── internal/
│   │   ├── config/config.go          # 环境配置
│   │   ├── model/                    # 数据模型
│   │   │   ├── user.go
│   │   │   ├── interest.go
│   │   │   ├── match.go
│   │   │   └── message.go
│   │   ├── repository/               # 数据访问
│   │   │   ├── user_repo.go
│   │   │   ├── match_repo.go
│   │   │   ├── chat_repo.go
│   │   │   └── interest_repo.go
│   │   ├── service/                  # 业务逻辑
│   │   │   ├── auth_service.go       # JWT 鉴权
│   │   │   ├── match_service.go      # 匹配算法
│   │   │   ├── chat_service.go       # 聊天服务
│   │   │   └── ws_service.go         # WebSocket Hub
│   │   ├── handler/                  # HTTP/WS 处理器
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── match_handler.go
│   │   │   ├── chat_handler.go
│   │   │   └── ws_handler.go
│   │   └── middleware/
│   │       └── auth_middleware.go     # JWT 中间件
│   ├── migrations/001_init.sql       # 数据库迁移 + 种子数据
│   └── Makefile
│
└── flutter/                          # Flutter 前端
    ├── pubspec.yaml
    └── lib/
        ├── main.dart
        ├── core/
        │   ├── theme/app_theme.dart   # 暗色主题
        │   ├── router/app_router.dart # 路由
        │   ├── network/api_client.dart# Dio HTTP 客户端
        │   └── constants/app_constants.dart
        ├── data/
        │   ├── models/               # JSON 模型
        │   │   ├── user_model.dart
        │   │   ├── match_model.dart
        │   │   └── message_model.dart
        │   └── providers/            # Riverpod 状态管理
        │       ├── auth_provider.dart
        │       ├── match_provider.dart
        │       └── chat_provider.dart
        └── presentation/
            ├── screens/              # 页面
            │   ├── login_screen.dart
            │   ├── register_screen.dart
            │   ├── personality_quiz_screen.dart
            │   ├── home_screen.dart
            │   ├── match_screen.dart
            │   ├── chat_list_screen.dart
            │   ├── chat_screen.dart
            │   └── profile_screen.dart
            └── widgets/              # 组件
                ├── match_card.dart
                └── message_bubble.dart
```

---

## 3. 数据库设计（10 张表）

```sql
users                    -- 用户表
interest_tags            -- 兴趣标签（80+ 个种子数据，8 大分类 × 平均 10 个）
user_interests           -- 用户-兴趣多对多
personality_questions    -- 大五人格问题（10 题）
personality_options      -- 每题选项（1-5 Likert）
user_personality_answers -- 用户答题记录
avatar_components        -- 虚拟形象组件（按拼装架构设计，V1.0 UI 仅暴露预设风格选择）
matches                  -- 用户匹配记录
chat_rooms               -- 聊天室（含 user_a_id + user_b_id，V1.0 仅 1 对 1）
messages                 -- 消息（含 client_msg_id 幂等）
device_tokens            -- 推送 Token
```

---

## 4. API 接口

### 鉴权（无需 Token）
```
POST /api/v1/auth/register    # 注册
POST /api/v1/auth/login       # 登录
```

### 用户（需要 Bearer Token）
```
GET    /api/v1/user/profile              # 获取个人信息
PUT    /api/v1/user/profile              # 更新个人信息
GET    /api/v1/user/tags                 # 兴趣标签列表
PUT    /api/v1/user/interests            # 保存我的兴趣
GET    /api/v1/user/personality/questions # 人格测试题目
POST   /api/v1/user/personality          # 提交人格测试
GET    /api/v1/user/personality          # 我的人格维度
GET    /api/v1/user/avatars              # 虚拟形象组件
POST   /api/v1/user/device-token        # 注册推送 Token
```

### 匹配
```
GET    /api/v1/match/candidates  # 推荐用户列表
POST   /api/v1/match/swipe       # 左滑(pass) / 右滑(like)
GET    /api/v1/match/list        # 已匹配列表
```

### 聊天
```
GET    /api/v1/chat/rooms                  # 聊天室列表
GET    /api/v1/chat/rooms/:id/messages     # 历史消息（支持分页）
POST   /api/v1/chat/rooms/:id/read         # 标记已读
```

### WebSocket
```
GET  /ws?token=xxx    # WebSocket 连接
```

**消息类型：**

| Type | 方向 | 说明 |
|---|---|---|
| `chat.message.send` | C→S | 发送消息 |
| `chat.message.ack` | S→C | 服务器确认 |
| `chat.message.new` | S→C | 新消息推送 |
| `chat.message.sync` | S→C | 离线消息同步 |
| `match.new` | S→C | 新匹配通知 |
| `system.heartbeat` | C↔S | 心跳 |
| `system.error` | S→C | 错误 |

---

## 5. 匹配算法

```
综合分 = 0.40 × Jaccard(兴趣重叠度)
       + 0.30 × Euclidean(人格向量距离，归一化)
       + 0.20 × 活跃度加成（1h内=1.0, 24h内=0.8, 3天内=0.5, 7天内=0.3, 其他=0.1）
       + 0.10 × 多样性奖励（发现新兴趣）
```

- 已匹配/已拒绝的不再出现
- 每天最多推荐 100 人
- 按综合分降序排列

---

## 6. WebSocket 可靠性设计

- **幂等发送**：每条消息带 `client_msg_id` (UUID)，数据库唯一索引防重复
- **离线队列**：接收方不在线时，消息存入 Redis List，7天 TTL（MVP 仅 Redis List）
- **重连同步**：客户端重连后，服务端自动推送离线消息
- **心跳保活**：30s 服务端 Ping，60s 客户端读超时
- **推送降级**：在线优先 WebSocket（200ms）→ 离线走 JPush（10s）→ JPush 失败降级 SMS（仅匹配通知）
- **水平扩展预留**：当前 Hub 为内存广播（单机），消息发送接口预留 MQ 转发点，日活破 5000 后切换为分布式消息队列广播
- **压测目标**：单机 WebSocket 5000 连接稳定，CPU < 70%，内存 < 2GB

---

## 7. 虚拟形象数据模型说明

V1.0 数据模型按**拼装架构**设计（`avatar_components` 表存组件级数据），但 UI 层仅暴露"选整套预设风格"的交互。V1.2 扩展到轻量拼装时，无需迁移数据模型，只需开放组件级选择 UI。

`chat_rooms` 表直接存储 `user_a_id` + `user_b_id`，`getReceiver()` 通过 sender_id 反查 room 中另一方即可，无需 JOIN 查询。

---

## 8. 人格测试

基于大五人格（Big Five）简化版，10 道选择题：

| 维度 | 题目数 | 说明 |
|---|---|---|
| 外向性 (Extraversion) | 2 | 社交能量来源 |
| 宜人性 (Agreeableness) | 2 | 合作与共情倾向 |
| 尽责性 (Conscientiousness) | 2 | 组织与自律程度 |
| 神经质 (Neuroticism) | 2 | 情绪稳定性 |
| 开放性 (Openness) | 2 | 对新事物的态度 |

每人 5 个维度各得一个 1-5 的均分，用于匹配计算。

---

## 9. 启动步骤

```bash
# 环境要求
# - Go 1.22+
# - Flutter 3.x+
# - PostgreSQL 15+
# - Redis 7+

# 1. 安装后端依赖
cd spark-app/backend
go mod tidy

# 2. 创建数据库并执行迁移
createdb spark
psql -d spark -f migrations/001_init.sql

# 3. 配置环境变量
export DB_HOST=localhost DB_PORT=5432 DB_USER=spark DB_PASSWORD=spark123 DB_NAME=spark
export REDIS_ADDR=localhost:6379
export JWT_SECRET=your-secret-key-change-in-prod

# 4. 启动后端
make run
# 或: go run cmd/server/main.go

# 5. 启动前端
cd ../flutter
flutter pub get
flutter run
```

---

## 10. 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `SERVER_PORT` | `8080` | 后端端口 |
| `DB_HOST` | `localhost` | 数据库地址 |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_USER` | `spark` | 数据库用户 |
| `DB_PASSWORD` | `spark123` | 数据库密码 |
| `DB_NAME` | `spark` | 数据库名 |
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `REDIS_PASS` | `` | Redis 密码 |
| `JWT_SECRET` | `spark-dev-secret-change-in-prod` | JWT 密钥 |

---

## 11. 开发计划（12 周，4 Phase）

| Phase | 内容 | 时间 |
|---|---|---|
| Phase 0 | 项目脚手架、JWT 认证、DB 迁移、微信 SDK + 实名认证 PoC、AI 选型 PoC、ICP 备案启动 | W1 |
| Phase 1 | Soul 式注册引导（标签选择 + 人格测试 + 形象预设）、埋点 SDK 集成、账号注销 | W2-3 |
| Phase 2 | 四维匹配引擎 + 滑动 UI、LBS 同城、星座适配报告、Bot 数据生成、种子用户招募 | W4-6 |
| Phase 3 | WebSocket 实时聊天、AI 破冰话题、JPush 推送、敏感词过滤（阿里云内容安全） | W7-9 |
| Phase 4 | 压测（WS 5000 连接）、性能优化、内测（100 人种子用户）、TestFlight + 应用商店提交 | W10-12 |

---

## 12. 敏感词过滤方案

采用**阿里云内容安全 API + 自建行业词库**混合方案：

| 层级 | 方案 | 说明 |
|---|---|---|
| 实时层 | 阿里云内容安全 API（发消息时调用） | 覆盖色情、暴恐、涉政、广告、辱骂 |
| 自建词库 | DFA 算法 + 变体词库 | 补充拼音变体、emoji 谐音、拆字等绕过手段 |
| T+1 层 | 阿里云内容安全异步审核 | 识别隐晦违规，日级扫描 |
| 人工层 | 举报触发 → 工单系统 → 24h SLA 处置 | 运营 W6 到岗后培训 |

成本估算（DAU 1000，人均日发 5 条，单价 0.0015 元/条）：月成本约 225 元。

---

## 13. 埋点与数据看板

- **埋点 SDK**：神策数据（免费版 100 万事件/月），Phase 1 W2 集成
- **核心漏斗**：app_open → browse_candidates → swipe_right → match → chat_open → message_send → message_reply
- **看板**：神策自带 + Metabase 自建，Phase 4 W11 搭建
- **预警**：次日留存 < 30% 企业微信告警，7 日留存 < 20% 邮件告警
- **周报**：PM 每周一输出数据周报（北极星指标趋势 + 9 项 MVP 指标 + 3 项归因 + 3 项行动项）

---

## 14. QA 策略

- **开发期（W1-W6）**：后端 2 人互相 code review，关键路径（auth、match、chat）写单元测试
- **测试期（W7-W12）**：外包测试公司（Testbird/Alltesting），覆盖功能测试 + 兼容性测试（iOS/Android 各 5 款机型）+ 性能测试 + **弱网场景测试（电梯/地铁，消息去重+离线合并）**
- **内测期（W11-W12）**：100 名种子用户内测，PM 收集反馈

---

## 15. AI 降级开关设计

所有 AI 功能（人格报告、星座适配、破冰话题、运势卡片）统一通过配置级开关控制：

- **存储**：Redis Key `ai_service_enabled`（`true` / `false`），Phase 0 W1 设计，Phase 1 W2 实现
- **后端逻辑**：每个 AI 功能调用前读取开关 → `true` 调 LLM API → `false` 读模板库 → 返回统一结构
- **前端不变**：前端 UI 接口不受影响，无论后端走 LLM 还是模板，返回的数据结构一致
- **切换场景**：LLM API 涨价/限流/宕机 → 运营在 admin 后台或直接改 Redis → 秒级切换，无需重发版
- **模板库**：每个 AI 功能预置 ≥ 20 条模板文案，覆盖主要场景，确保模板模式下体验不明显降级

---

## 16. Flutter WebSocket 弱网策略

不从头手写 WS 底层逻辑，基于成熟库封装：

- **基础库**：`web_socket_channel` + `connectivity_plus`（网络状态监听）
- **重连策略**：指数退避（1s → 2s → 4s → 8s → 30s 上限），App 切前台立即重连
- **消息去重**：`client_msg_id`（UUID）在客户端本地 SQLite 缓存 + 服务端唯一索引，双保险
- **离线合并**：重连后拉取 Redis 离线队列，按时间戳合并到本地消息列表，去重后展示
- **弱网测试**：W8 QA 模拟电梯/地铁场景（丢包 10-30%、延迟 500-2000ms、断连 5-30s），验证消息不丢不重
- **前端精力分配**：W7 1 人封装 WS 层（2 天），另 1 人写聊天 UI（3 天），W8 集中搞弱网测试

---

## 17. 虚拟形象资产策略

V1.0 不自绘虚拟形象，使用现成资产：

- **方案 A（推荐）**：购买 2D/3D 虚拟形象资产授权（如 Ready Player Me、Ready Maker Studio），6-8 套风格，单价约 $50-200/套
- **方案 B**：使用 AI 生成风格化头像（Midjourney/Stable Diffusion 批量生成），统一风格后作为预设
- **W1 决策**：设计师 W1 D4-5 评估方案 A vs B，选定后输出 6-8 套形象
- **V1.2 扩展**：若后续支持轻量拼装，需重新评估资产格式兼容性

---

## 18. 团队配置

- Flutter 开发 × 2（全职，W1 到岗）
- Go 后端开发 × 2（全职，W1 到岗）
- UI/UX 设计师 × 1（全职，W1 到岗）
- 产品经理 × 1（全职，兼 W1-W6 运营）
- QA 测试 × 外包（Testbird/Alltesting，W7 启动）
- 运营/增长 × 1（全职，W6 到岗）
