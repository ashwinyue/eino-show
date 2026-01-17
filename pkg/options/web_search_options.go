// Package options 提供 Web 搜索配置选项
package options

import "github.com/spf13/pflag"

// WebSearchProviderConfig Web 搜索提供商配置
type WebSearchProviderConfig struct {
	ID             string `json:"id" mapstructure:"id"`                         // 提供商ID
	Name           string `json:"name" mapstructure:"name"`                     // 提供商名称
	Free           bool   `json:"free" mapstructure:"free"`                     // 是否免费
	RequiresAPIKey bool   `json:"requires_api_key" mapstructure:"requires_api_key"` // 是否需要API密钥
	Description    string `json:"description" mapstructure:"description"`       // 描述
	APIURL         string `json:"api_url" mapstructure:"api_url"`               // API地址（可选）
}

// WebSearchDefaultConfig Web 搜索默认配置
type WebSearchDefaultConfig struct {
	Provider          string   `json:"provider" mapstructure:"provider"`                 // 默认提供商ID
	MaxResults        int      `json:"max_results" mapstructure:"max_results"`           // 最大搜索结果数
	IncludeDate       bool     `json:"include_date" mapstructure:"include_date"`         // 是否包含日期
	CompressionMethod string   `json:"compression_method" mapstructure:"compression_method"` // 压缩方法
	Blacklist         []string `json:"blacklist" mapstructure:"blacklist"`               // 黑名单规则列表
}

// WebSearchOptions Web 搜索配置选项
type WebSearchOptions struct {
	Providers []WebSearchProviderConfig `json:"providers" mapstructure:"providers"` // 可用搜索引擎列表
	Default   WebSearchDefaultConfig    `json:"default" mapstructure:"default"`     // 默认配置
	Timeout   int                       `json:"timeout" mapstructure:"timeout"`     // 超时时间（秒）
	ProxyURL  string                    `json:"proxy_url" mapstructure:"proxy_url"` // 代理地址
}

// NewWebSearchOptions 创建默认的 Web 搜索配置.
func NewWebSearchOptions() *WebSearchOptions {
	return &WebSearchOptions{
		Providers: []WebSearchProviderConfig{
			{
				ID:             "duckduckgo",
				Name:           "DuckDuckGo",
				Free:           true,
				RequiresAPIKey: false,
				Description:    "DuckDuckGo 搜索引擎，无需 API Key",
			},
		},
		Default: WebSearchDefaultConfig{
			Provider:          "duckduckgo",
			MaxResults:        5,
			IncludeDate:       true,
			CompressionMethod: "none",
			Blacklist:         []string{},
		},
		Timeout: 10,
	}
}

// Validate 验证 Web 搜索配置.
func (o *WebSearchOptions) Validate() []error {
	return nil
}

// AddFlags 添加 Web 搜索相关标志（主要用于从配置文件读取）
func (o *WebSearchOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	// Web 搜索配置主要通过配置文件设置，这里仅添加基本的超时和代理设置
	fs.IntVar(&o.Timeout, fullPrefix+".timeout", o.Timeout, "Web search timeout in seconds.")
	fs.StringVar(&o.ProxyURL, fullPrefix+".proxy-url", o.ProxyURL, "Proxy URL for web search requests.")
}
