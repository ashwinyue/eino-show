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

    // 通过 ID 更新 qwen-turbo 的 API Key
    MODEL_ID := "c4227a56-9d5a-4780-ad2c-8e4e0180f218"
    newParams := `{"base_url":"https://dashscope.aliyuncs.com/api/v1","api_key":"sk-a6aaf8beba18425e9942c4a33ae58caf"}`
    
    result := db.Exec("UPDATE models SET parameters = ? WHERE id = ?", newParams, MODEL_ID)
    
    if result.Error != nil {
        log.Fatalf("Update failed: %v", result.Error)
    }
    if result.RowsAffected == 0 {
        log.Fatal("No rows affected")
    }

    fmt.Printf("Model %s API Key updated successfully!\n", MODEL_ID)
}
