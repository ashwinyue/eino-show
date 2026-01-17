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

    // 执行 SQL 创建表
    sql := `
    CREATE TABLE IF NOT EXISTS models (
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
    );

    CREATE INDEX IF NOT EXISTS idx_models_type ON models(type);
    CREATE INDEX IF NOT EXISTS idx_models_source ON models(source);
    CREATE INDEX IF NOT EXISTS idx_models_is_default ON models(is_default);
    CREATE INDEX IF NOT EXISTS idx_models_tenant_id ON models(tenant_id);
    CREATE INDEX IF NOT EXISTS idx_models_deleted_at ON models(deleted_at);
    `

    if err := db.Exec(sql).Error; err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }

    // 检查是否有数据，没有则插入默认数据
    var count int64
    db.Raw("SELECT COUNT(*) FROM models").Scan(&count)
    
    if count == 0 {
        fmt.Println("Inserting default models...")
        
        // gpt-4o-mini
        db.Exec(`INSERT INTO models (name, type, source, description, parameters, is_default, status, is_builtin, tenant_id) 
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)`,
            "gpt-4o-mini", "llm", "openai", "OpenAI GPT-4o Mini", 
            `{"base_url":"https://api.openai.com/v1","api_key":"your-api-key"}`, 
            true, "active", true)

        // text-embedding-3-small  
        db.Exec(`INSERT INTO models (name, type, source, description, parameters, is_default, status, is_builtin, tenant_id) 
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)`,
            "text-embedding-3-small", "embedding", "openai", "OpenAI text-embedding-3-small",
            `{"base_url":"https://api.openai.com/v1","api_key":"your-api-key","dimension":1536}`,
            true, "active", true)

        fmt.Println("Default models inserted!")
    }

    fmt.Println("Models table created successfully!")
}
