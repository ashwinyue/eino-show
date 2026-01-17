// Package tool 提供 WebFetch 工具，用于抓取网页内容.
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// WebFetchToolName WebFetch 工具名称.
const WebFetchToolName = "web_fetch"

// WebFetchConfig WebFetch 工具配置.
type WebFetchConfig struct {
	// ChatModel LLM 模型（可选，用于分析网页内容）
	ChatModel model.ChatModel
	// MaxContentLength 最大内容长度（默认 100000 字节）
	MaxContentLength int
	// Timeout HTTP 超时（默认 60 秒）
	Timeout time.Duration
}

// NewWebFetchTool 创建 WebFetch 工具.
func NewWebFetchTool() tool.InvokableTool {
	return NewWebFetchToolWithConfig(nil)
}

// NewWebFetchToolWithConfig 使用配置创建 WebFetch 工具.
func NewWebFetchToolWithConfig(cfg *WebFetchConfig) tool.InvokableTool {
	t := &webFetchTool{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		maxContentLength: 100000,
	}

	if cfg != nil {
		t.chatModel = cfg.ChatModel
		if cfg.MaxContentLength > 0 {
			t.maxContentLength = cfg.MaxContentLength
		}
		if cfg.Timeout > 0 {
			t.client.Timeout = cfg.Timeout
		}
	}

	return t
}

// webFetchTool 网页抓取工具.
type webFetchTool struct {
	client           *http.Client
	chatModel        model.ChatModel
	maxContentLength int
}

// webFetchInput WebFetch 输入参数.
type webFetchInput struct {
	Items []webFetchItem `json:"items"`
}

// webFetchItem 单个抓取任务.
type webFetchItem struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// Info 返回工具信息.
func (t *webFetchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	_ = ctx
	return &schema.ToolInfo{
		Name: WebFetchToolName,
		Desc: `Fetch detailed web content from URLs and analyze it.

## When to Use
- After web_search returns results with truncated content
- When you need full page content to answer the question
- When web_search snippet is insufficient

## Parameters
- items: Array of {url, prompt} combinations
  - url: The webpage URL to fetch (should come from web_search results)
  - prompt: Analysis prompt for the page content

## Returns
- Summary result for each URL
- Original content fragment
- Error information if fetch fails`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"items": {
				Type:     "array",
				Desc:     "Array of fetch tasks with url and prompt",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具.
func (t *webFetchTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	// 解析参数
	var input webFetchInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if len(input.Items) == 0 {
		return "", fmt.Errorf("missing required parameter: items")
	}

	// 执行抓取
	var results []string
	for i, item := range input.Items {
		if err := t.validateURL(item.URL); err != nil {
			results = append(results, fmt.Sprintf("Item #%d:\nURL: %s\n错误: %v\n", i+1, item.URL, err))
			continue
		}

		output, err := t.fetchURL(ctx, item.URL, item.Prompt)
		if err != nil {
			results = append(results, fmt.Sprintf("Item #%d:\nURL: %s\n错误: %v\n", i+1, item.URL, err))
			continue
		}
		results = append(results, fmt.Sprintf("Item #%d:\n%s\n", i+1, output))
	}

	return fmt.Sprintf("=== Web Fetch Results ===\n\n%s", strings.Join(results, "\n")), nil
}

// validateURL 验证 URL.
func (t *webFetchTool) validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	return nil
}

// fetchURL 抓取 URL 内容.
func (t *webFetchTool) fetchURL(ctx context.Context, url, prompt string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; eino-show/1.0)")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// 限制读取大小
	limitedReader := io.LimitReader(resp.Body, int64(t.maxContentLength))
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 转换为纯文本
	textContent := t.htmlToText(string(content))

	// 如果有 LLM 模型，使用 LLM 分析内容
	if t.chatModel != nil && prompt != "" {
		analysis, err := t.analyzeWithLLM(ctx, url, textContent, prompt)
		if err != nil {
			// LLM 分析失败，回退到原始内容
			return t.formatRawOutput(url, prompt, textContent), nil
		}
		return analysis, nil
	}

	return t.formatRawOutput(url, prompt, textContent), nil
}

// analyzeWithLLM 使用 LLM 分析网页内容.
func (t *webFetchTool) analyzeWithLLM(ctx context.Context, url, content, prompt string) (string, error) {
	// 截断过长的内容以适应 LLM 上下文窗口
	maxChars := 8000
	if len(content) > maxChars {
		content = content[:maxChars] + "\n[Content truncated for analysis]"
	}

	systemPrompt := `You are a web content analyzer. Your task is to analyze the provided web page content and extract relevant information based on the user's prompt.

Guidelines:
1. Focus on information relevant to the user's prompt
2. Provide a clear and concise summary
3. Include specific facts, numbers, or quotes when available
4. If the content doesn't contain relevant information, state that clearly`

	userMessage := fmt.Sprintf(`## Web Page URL
%s

## User's Analysis Request
%s

## Web Page Content
%s

Please analyze the above content based on the user's request.`, url, prompt, content)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	}

	resp, err := t.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM analysis failed: %w", err)
	}

	return fmt.Sprintf("## Analysis Result for %s\n\n**Query:** %s\n\n%s", url, prompt, resp.Content), nil
}

// formatRawOutput 格式化原始输出.
func (t *webFetchTool) formatRawOutput(url, prompt, content string) string {
	// 截断过长的内容
	maxChars := 5000
	if len(content) > maxChars {
		content = content[:maxChars] + "...\n[Content truncated]"
	}

	output := fmt.Sprintf("## URL: %s\n", url)
	if prompt != "" {
		output += fmt.Sprintf("**Analysis Prompt:** %s\n", prompt)
	}
	output += fmt.Sprintf("\n### Content:\n%s\n", content)

	return output
}

// htmlToText 简单的 HTML 转 文本.
func (t *webFetchTool) htmlToText(htmlContent string) string {
	// 移除脚本
	htmlContent = removeBetween(htmlContent, "<script", "</script>")
	// 移除样式
	htmlContent = removeBetween(htmlContent, "<style", "</style>")

	// 移除 HTML 标签 - 简化版本
	result := strings.Builder{}
	inTag := false
	for _, r := range htmlContent {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	// 清理空白
	lines := strings.Split(result.String(), "\n")
	cleanLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// removeBetween 移除两个标记之间的内容.
func removeBetween(s, start, end string) string {
	result := s
	for {
		startIdx := strings.Index(result, start)
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(result[startIdx:], end)
		if endIdx == -1 {
			break
		}
		result = result[:startIdx] + result[startIdx+len(end)+endIdx:]
	}
	return result
}
