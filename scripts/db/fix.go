//go:build ignore

package main

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    dsn := "host=127.0.0.1 port=5432 user=einoshow password=einoshow1234 dbname=einoshow sslmode=disable"
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // 删除旧表
    db.Exec("DROP TABLE IF EXISTS models CASCADE")

    // 创建新表（完整的结构）
    db.Exec(`
        CREATE TABLE models (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            tenant_id INT NOT NULL,
            name VARCHAR(255) NOT NULL,
            type VARCHAR(50) NOT NULL,
            source VARCHAR(50) NOT NULL,
            description TEXT,
            parameters TEXT NOT NULL DEFAULT '{}',
            is_default BOOLEAN NOT NULL DEFAULT FALSE,
            status VARCHAR(50) NOT NULL DEFAULT 'active',
            is_builtin BOOLEAN NOT NULL DEFAULT FALSE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP
        )
    `)

    // 创建索引
    db.Exec("CREATE INDEX idx_models_type ON models(type)")
    db.Exec("CREATE INDEX idx_models_source ON models(source)")
    db.Exec("CREATE INDEX idx_models_is_default ON models(is_default)")
    db.Exec("CREATE INDEX idx_models_tenant_id ON models(tenant_id)")
    db.Exec("CREATE INDEX idx_models_deleted_at ON models(deleted_at)")

    // 插入默认模型
    desc1 := "OpenAI GPT-4o Mini"
    desc2 := "OpenAI text-embedding-3-small"
    
    db.Exec(`INSERT INTO models (name, type, source, description, parameters, is_default, status, is_builtin, tenant_id) 
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)`,
        "gpt-4o-mini", "llm", "openai", &desc1, 
        `{"base_url":"https://api.openai.com/v1","api_key":"your-api-key"}`, 
        true, "active", true)

    db.Exec(`INSERT INTO models (name, type, source, description, parameters, is_default, status, is_builtin, tenant_id) 
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)`,
        "text-embedding-3-small", "embedding", "openai", &desc2, 
        `{"base_url":"https://api.openai.com/v1","api_key":"your-api-key","dimension":1536}`, 
        true, "active", true)

    fmt.Println("Models table recreated successfully!")
}
