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

    // 查询所有模型
    type Model struct {
        ID          string
        Name        string
        Parameters  string
    }
    var models []Model
    db.Find(&models)

    fmt.Println("=== 当前所有模型 ===")
    for _, m := range models {
        fmt.Printf("ID: %v\n", m.ID)
        fmt.Printf("Name: %v\n", m.Name)
        fmt.Printf("Parameters: %v\n", m.Parameters)
        fmt.Println("---")
    }
}
