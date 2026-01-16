// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/onexstack/onexstack/pkg/db"
	"github.com/spf13/pflag"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// 帮助信息文本.
const helpText = `Usage: gen_gorm_model [flags]

Generate GORM model code from database schema.

Flags:
`

// Querier 定义了数据库查询接口.
type Querier interface {
	// FilterWithNameAndRole 按名称和角色查询记录
	FilterWithNameAndRole(name string) ([]gen.T, error)
}

// GenerateConfig 保存代码生成的配置.
type GenerateConfig struct {
	ModelPackagePath string
	GenerateFunc     func(g *gen.Generator)
}

// 预定义的生成配置.
var generateConfigs = map[string]GenerateConfig{
	"mb": {ModelPackagePath: "github.com/ashwinyue/eino-show/internal/apiserver/model", GenerateFunc: GenerateMiniBlogModels},
}

// 命令行参数.
var (
	dbType     = pflag.String("db-type", "postgresql", "Database type: mysql or postgresql")
	addr       = pflag.StringP("addr", "a", "127.0.0.1:5432", "Database host:port address.")
	username   = pflag.StringP("username", "u", "einoshow", "Database username.")
	password   = pflag.StringP("password", "p", "einoshow1234", "Database password.")
	database   = pflag.StringP("db", "d", "einoshow", "Database name.")
	sslMode    = pflag.String("ssl-mode", "disable", "PostgreSQL SSL mode (disable, require, verify-ca, verify-full).")
	modelPath  = pflag.String("model-pkg-path", "", "Generated model code's package path.")
	components = pflag.StringSlice("component", []string{"mb"}, "Generated model code's for specified component.")
	help       = pflag.BoolP("help", "h", false, "Show this help message.")
)

func main() {
	// 设置自定义的使用说明函数
	pflag.Usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
	pflag.Parse()

	// 如果设置了帮助标志，则显示帮助信息并退出
	if *help {
		pflag.Usage()
		return
	}

	// 初始化数据库连接
	dbInstance, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 处理组件并生成代码
	for _, component := range *components {
		processComponent(component, dbInstance)
	}
}

// initializeDatabase 创建并返回一个数据库连接.
func initializeDatabase() (*gorm.DB, error) {
	switch *dbType {
	case "mysql":
		return connectMySQL()
	case "postgresql", "postgres":
		return connectPostgreSQL()
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: mysql, postgresql)", *dbType)
	}
}

// connectMySQL 连接 MySQL 数据库.
func connectMySQL() (*gorm.DB, error) {
	dbOptions := &db.MySQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *database,
	}
	return db.NewMySQL(dbOptions)
}

// connectPostgreSQL 连接 PostgreSQL 数据库.
func connectPostgreSQL() (*gorm.DB, error) {
	dbOptions := &db.PostgreSQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *database,
		SSLMode:  *sslMode,
	}
	return db.NewPostgreSQL(dbOptions)
}

// processComponent 处理单个组件以生成代码.
func processComponent(component string, dbInstance *gorm.DB) {
	config, ok := generateConfigs[component]
	if !ok {
		log.Printf("Component '%s' not found in configuration. Skipping.", component)
		return
	}

	// 解析模型包路径
	modelPkgPath := resolveModelPackagePath(config.ModelPackagePath)

	// 创建生成器实例
	generator := createGenerator(modelPkgPath)
	generator.UseDB(dbInstance)

	// 应用自定义生成器选项
	applyGeneratorOptions(generator)

	// 使用指定的函数生成模型
	config.GenerateFunc(generator)

	// 执行代码生成
	generator.Execute()
}

// resolveModelPackagePath 确定模型生成的包路径.
func resolveModelPackagePath(defaultPath string) string {
	if *modelPath != "" {
		return *modelPath
	}
	// 如果是 Go 模块路径，转换为绝对路径
	if filepath.IsAbs(defaultPath) {
		return defaultPath
	}
	// 处理 Go 模块路径格式 (github.com/...)
	return defaultPath
}

// createGenerator 初始化并返回一个新的生成器实例.
func createGenerator(packagePath string) *gen.Generator {
	return gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		ModelPkgPath:      packagePath,
		WithUnitTest:      true,
		FieldNullable:     true,  // 对于数据库中可空的字段，使用指针类型。
		FieldSignable:     false, // 禁用无符号属性以提高兼容性。
		FieldWithIndexTag: false, // 不包含 GORM 的索引标签。
		FieldWithTypeTag:  false, // 不包含 GORM 的类型标签。
	})
}

// applyGeneratorOptions 设置自定义生成器选项.
func applyGeneratorOptions(g *gen.Generator) {
	// 为特定字段自定义 GORM 标签
	g.WithOpts(
		gen.FieldGORMTag("createdAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
		gen.FieldGORMTag("updatedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "current_timestamp")
			return tag
		}),
	)
}

// GenerateMiniBlogModels 为 miniblog 组件生成模型.
func GenerateMiniBlogModels(g *gen.Generator) {
	// Casbin 规则模型
	g.GenerateModelAs(
		"casbin_rule",
		"CasbinRuleM",
		gen.FieldRename("ptype", "PType"),
		gen.FieldIgnore("placeholder"),
	)

	// ========== 用户系统 ==========

	// 用户模型（表名: users）
	g.GenerateModelAs(
		"users",
		"UserM",
	)

	// ========== Agent 系统 ==========

	// 自定义 Agent 模型
	g.GenerateModelAs(
		"custom_agents",
		"CustomAgentM",
	)

	// MCP 服务模型
	g.GenerateModelAs(
		"mcp_services",
		"MCPServiceM",
	)

	// ========== Session 系统 ==========

	// 会话模型
	g.GenerateModelAs(
		"sessions",
		"SessionM",
	)

	// 会话项模型
	g.GenerateModelAs(
		"session_items",
		"SessionItemM",
	)

	// 消息模型
	g.GenerateModelAs(
		"messages",
		"MessageM",
	)

	// 认证令牌模型
	g.GenerateModelAs(
		"auth_tokens",
		"AuthTokenM",
	)

	// ========== 知识库系统 ==========

	// 知识库模型
	g.GenerateModelAs(
		"knowledge_bases",
		"KnowledgeBaseM",
	)

	// 知识项模型（文档）
	g.GenerateModelAs(
		"knowledges",
		"KnowledgeM",
	)

	// 知识分块模型（向量存储）
	g.GenerateModelAs(
		"chunks",
		"ChunkM",
	)

	// 知识标签模型
	g.GenerateModelAs(
		"knowledge_tags",
		"KnowledgeTagM",
	)

	// ========== 租户系统 ==========

	// 租户模型
	g.GenerateModelAs(
		"tenants",
		"TenantM",
	)

	// 模型配置
	g.GenerateModelAs(
		"models",
		"LLMModelM",
	)
}
