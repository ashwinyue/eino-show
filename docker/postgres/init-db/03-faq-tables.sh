#!/bin/bash
# FAQ Tables Migration Script
# Run this script to create FAQ related tables

set -e

echo "[INFO] Starting FAQ tables migration..."

psql -v ON_ERROR_STOP=1 \
    --username "$POSTGRES_USER" \
    --dbname "$POSTGRES_DB" \
    -f /docker-entrypoint-initdb.d/03-faq-tables.sql

echo "[SUCCESS] FAQ tables migration completed!"
