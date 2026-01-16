# eino-show Docker 开发环境

## 目录结构

```
docker/
├── postgres/
│   └── init-db/
│       ├── 00-init-extensions.sh      # ParadeDB 扩展初始化
│       ├── 01-weknora-migration.sh     # WeKnora 数据迁移脚本
│       └── weknora-migration.sql       # WeKnora 数据库导出（已清理）
└── README.md
```

## 服务说明

### PostgreSQL (ParadeDB)
- **镜像**: `paradedb/paradedb:v0.18.9-pg17`
- **端口**: `5432`
- **特性**:
  - 内置 pgvector 扩展（向量搜索）
  - 内置 ParadeDB 扩展（全文搜索）
- **数据迁移**: 首次启动时自动从 WeKnora 迁移数据

### Redis
- **镜像**: `redis:7.0-alpine`
- **端口**: `6379`
- **持久化**: AOF 模式

## 使用方法

### 首次启动（迁移 WeKnora 数据）

```bash
make dev-start
# 或
./scripts/dev.sh start
```

首次启动会自动：
1. 创建 PostgreSQL 和 Redis 容器
2. 初始化 ParadeDB 扩展
3. 导入 WeKnora 数据（约 8MB，包含表结构和数据）

### 查看服务状态

```bash
make dev-status
# 或
./scripts/dev.sh status
```

### 查看日志

```bash
make dev-logs
# 或
./scripts/dev.sh logs
```

### 停止服务

```bash
make dev-stop
# 或
./scripts/dev.sh stop
```

### 重启服务

```bash
make dev-restart
# 或
./scripts/dev.sh restart
```

## 数据迁移说明

### WeKnora 数据表

迁移包含以下表结构及数据：
- `sessions` - 会话记录
- `session_items` - 会话项
- `messages` - 消息记录
- `custom_agents` - 自定义 Agent
- `knowledge_bases` - 知识库
- `knowledges` - 知识条目
- `chunks` - 文本分块（含向量嵌入）
- `users` - 用户
- `tenants` - 租户

### SQL 文件说明

原始文件: `docs/sql/public.sql` (7.9MB)
迁移文件: `docker/postgres/init-db/weknora-migration.sql` (8.2MB)

清理内容:
- 移除 `OWNER TO` 语句（避免权限冲突）
- 保留所有数据定义和 INSERT 语句

## 环境配置

编辑 `.env` 文件配置数据库连接：

```bash
# PostgreSQL
DB_PORT=5432
DB_USER=einoshow
DB_PASSWORD=einoshow1234
DB_NAME=einoshow

# Redis
REDIS_PORT=6379
REDIS_PASSWORD=
```

## 故障排查

### 数据库连接失败
```bash
# 检查容器状态
docker ps | grep eino-show

# 查看 PostgreSQL 日志
docker logs eino-show-postgres-dev
```

### 重新导入数据
```bash
# 1. 停止并删除容器和数据卷
make dev-stop
docker volume rm eino-show_postgres-data-dev

# 2. 重新启动（会触发初始化）
make dev-start
```

### 进入数据库
```bash
# 使用 psql 连接
docker exec -it eino-show-postgres-dev psql -U einoshow -d einoshow
```
