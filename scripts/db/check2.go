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

    // 查看表结构
    fmt.Println("=== 查看所有模型 ===")
    var results []map[string]any
    db.Table("models").Find(&results)
    for _, r := range results {
        fmt.Printf("ID: %v, Name: %v\n", r["id"], r["name"])
    }
}
