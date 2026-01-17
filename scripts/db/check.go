//go:build ignore

package main

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type LLMModelM struct {
    ID          string
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

    var model LLMModelM
    result := db.Where("name = ?", "qwen-turbo").First(&model)
    if result.Error != nil {
        log.Fatalf("Query failed: %v", result.Error)
    }

    fmt.Printf("Model ID: %s\n", model.ID)
    fmt.Printf("Name: %s\n", model.Name)
    fmt.Printf("Parameters: %s\n", model.Parameters)
}
