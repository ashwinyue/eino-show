#!/bin/bash
# WeKnora 数据迁移脚本
# 使用 ON_ERROR_STOP=0 容忍类型冲突等错误

set +e  # 不在错误时退出

echo "[INFO] 开始迁移 WeKnora 数据..."

psql -v ON_ERROR_STOP=0 \
    --username "$POSTGRES_USER" \
    --dbname "$POSTGRES_DB" \
    -f /docker-entrypoint-initdb.d/02-weknora-migration-data.sql.bak

echo "[SUCCESS] WeKnora 数据迁移完成（部分类型错误可忽略）"
