package options

// LLMOptions LLM 模型配置（用于内置 Agent）.
type LLMOptions struct {
	Provider string `json:"provider" mapstructure:"provider"` // 模型提供商 (openai, ark, dashscope 等)
	Model    string `json:"model" mapstructure:"model"`       // 模型名称
	APIKey   string `json:"api_key" mapstructure:"api_key"`   // API 密钥
	BaseURL  string `json:"base_url" mapstructure:"base_url"` // API 基础 URL
}

// NewLLMOptions 创建默认的 LLM 配置.
func NewLLMOptions() *LLMOptions {
	return &LLMOptions{
		Provider: "openai",
		Model:    "gpt-4o-mini",
	}
}

// Validate 验证 LLM 配置.
func (o *LLMOptions) Validate() []error {
	var errs []error
	// API Key 和 BaseURL 可以从环境变量获取，所以这里不强制校验
	return errs
}
