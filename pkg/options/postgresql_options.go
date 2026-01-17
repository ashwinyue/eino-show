package options

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/pflag"
	"gorm.io/gorm"

	"github.com/ashwinyue/eino-show/pkg/db"
	gormlogger "github.com/ashwinyue/eino-show/pkg/logger/slog/gorm"
)

var _ IOptions = (*PostgreSQLOptions)(nil)

// PostgreSQLOptions 定义 PostgreSQL 数据库配置选项.
type PostgreSQLOptions struct {
	Addr                  string        `json:"addr,omitempty" mapstructure:"addr"`
	Username              string        `json:"username,omitempty" mapstructure:"username"`
	Password              string        `json:"-" mapstructure:"password"`
	Database              string        `json:"database" mapstructure:"database"`
	SSLMode               string        `json:"ssl-mode" mapstructure:"ssl-mode"`
	MaxIdleConnections    int           `json:"max-idle-connections,omitempty" mapstructure:"max-idle-connections,omitempty"`
	MaxOpenConnections    int           `json:"max-open-connections,omitempty" mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `json:"max-connection-life-time,omitempty" mapstructure:"max-connection-life-time"`
	LogLevel              int           `json:"log-level" mapstructure:"log-level"`
}

// NewPostgreSQLOptions 创建一个零值实例.
func NewPostgreSQLOptions() *PostgreSQLOptions {
	return &PostgreSQLOptions{
		Addr:                  "127.0.0.1:5432",
		Username:              "",
		Password:              "",
		Database:              "",
		SSLMode:               "disable",
		MaxIdleConnections:    100,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Duration(10) * time.Second,
		LogLevel:              1, // Silent
	}
}

// Validate 验证 PostgreSQLOptions 的参数.
func (o *PostgreSQLOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags 为 PostgreSQL 相关的配置添加命令行标志.
func (o *PostgreSQLOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr, ""+
		"PostgreSQL service host address.")
	fs.StringVar(&o.Username, fullPrefix+".username", o.Username, "Username for access to PostgreSQL service.")
	fs.StringVar(&o.Password, fullPrefix+".password", o.Password, ""+
		"Password for access to PostgreSQL, should be used pair with password.")
	fs.StringVar(&o.Database, fullPrefix+".database", o.Database, ""+
		"Database name for the server to use.")
	fs.StringVar(&o.SSLMode, fullPrefix+".ssl-mode", o.SSLMode, ""+
		"SSL mode for PostgreSQL connection (disable, require, verify-ca, verify-full).")
	fs.IntVar(&o.MaxIdleConnections, fullPrefix+".max-idle-connections", o.MaxOpenConnections, ""+
		"Maximum idle connections allowed to connect to PostgreSQL.")
	fs.IntVar(&o.MaxOpenConnections, fullPrefix+".max-open-connections", o.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to PostgreSQL.")
	fs.DurationVar(&o.MaxConnectionLifeTime, fullPrefix+".max-connection-life-time", o.MaxConnectionLifeTime, ""+
		"Maximum connection life time allowed to connect to PostgreSQL.")
	fs.IntVar(&o.LogLevel, fullPrefix+".log-mode", o.LogLevel, ""+
		"Specify gorm log level.")
}

// DSN 返回 PostgreSQL 连接字符串.
func (o *PostgreSQLOptions) DSN() string {
	return fmt.Sprintf(`host=%s user=%s password=%s dbname=%s sslmode=%s`,
		o.Addr,
		o.Username,
		o.Password,
		o.Database,
		o.SSLMode,
	)
}

// NewDB 创建 PostgreSQL 数据库连接.
func (o *PostgreSQLOptions) NewDB() (*gorm.DB, error) {
	opts := &db.PostgreSQLOptions{
		Addr:                  o.Addr,
		Username:              o.Username,
		Password:              o.Password,
		Database:              o.Database,
		SSLMode:               o.SSLMode,
		MaxIdleConnections:    o.MaxIdleConnections,
		MaxOpenConnections:    o.MaxOpenConnections,
		MaxConnectionLifeTime: o.MaxConnectionLifeTime,
		Logger:                gormlogger.New(slog.Default()),
	}

	return db.NewPostgreSQL(opts)
}
