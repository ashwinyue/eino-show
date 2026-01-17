//go:build ignore

package main

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type LLMModelM struct {
    ID          string  `gorm:"primaryKey"`
    TenantID    int32
    Name        string
    Type        string
    Source      string
    Description *string
    Parameters  string
    IsDefault   bool
    Status      string
    IsBuiltin   bool
}

func (LLMModelM) TableName() string {
    return "models"
}

func main() {
    dsn := "host=127.0.0.1 port=5432 user=einoshow password=einoshow1234 dbname=einoshow sslmode=disable"
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // 创建表
    if err := db.AutoMigrate(&LLMModelM{}); err != nil {
        log.Fatalf("Failed to migrate: %v", err)
    }

    // 插入默认模型
    desc := "OpenAI GPT-4o Mini"
    db.Create(&LLMModelM{
        Name:        "gpt-4o-mini",
        Type:        "llm",
        Source:      "openai",
        Description: &desc,
        Parameters:  `{"base_url":"https://api.openai.com/v1","api_key":"your-api-key"}`,
        IsDefault:   true,
        Status:      "active",
        IsBuiltin:   true,
        TenantID:    0,
    })

    fmt.Println("Models table created successfully!")
}
