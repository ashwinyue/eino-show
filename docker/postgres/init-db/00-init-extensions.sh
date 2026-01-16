#!/bin/bash
# eino-show PostgreSQL 初始化脚本
# 确保必要的扩展已启用

set -e

echo "[INFO] 初始化 PostgreSQL 扩展..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- UUID 生成器扩展 (WeKnora 数据需要)
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    -- pgvector 扩展 (ParadeDB 内置，确认启用)
    SELECT * FROM pg_extension WHERE extname = 'vector';

    -- ParadeDB 扩展确认
    SELECT * FROM pg_extension WHERE extname = 'paradedb';

    -- 显示版本信息
    SELECT version();
EOSQL

echo "[SUCCESS] PostgreSQL 扩展初始化完成"
