# eino-show 项目文档

## 目录结构

| 目录 | 说明 |
|------|------|
| `api/` | API 接口文档（参考 WeKnora，用于兼容性验证） |
| `devel/` | 开发规范和约定 |
| `guide/` | 用户指南 |
| `images/` | 文档图片 |
| `plan/` | 重构计划 |
| `reference/` | 参考文档（从 WeKnora 迁移） |

## 快速导航

### 开发相关
- [开发规约](../CONVENTION.md) - 项目开发规范
- [Claude 指南](../CLAUDE.md) - AI 开发助手指南
- [重构计划](./plan/重构计划.md) - 分阶段实施计划

### API 文档
- [API 概览](./api/README.md) - 接口总览
- [会话管理](./api/session.md) - Session API
- [知识库管理](./api/knowledge-base.md) - Knowledge API
- [聊天功能](./api/chat.md) - Chat API

### 参考文档
- [Eino ADK 重构设计](./reference/eino-adk-redesign.md) - 基于 Eino ADK 的重构设计
- [Agent RAG 开发指南](./reference/AGENT_RAG_DEVELOPMENT_GUIDE.md) - RAG 开发
- [内置模型管理](./reference/BUILTIN_MODELS.md) - 模型配置
