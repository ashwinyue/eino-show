// Package tool 提供 Eino 工具实现.
package tool

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AliasedTool 工具别名包装器.
// 允许一个工具以不同的名称注册，解决 LLM 调用错误工具名的问题.
type AliasedTool struct {
	wrapped tool.InvokableTool
	alias   string
}

// NewAliasedTool 创建工具别名包装器.
func NewAliasedTool(t tool.InvokableTool, alias string) *AliasedTool {
	return &AliasedTool{
		wrapped: t,
		alias:   alias,
	}
}

// 确保 AliasedTool 实现了 InvokableTool 接口.
var _ tool.InvokableTool = (*AliasedTool)(nil)

// Info 返回工具信息（使用别名）.
func (t *AliasedTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	info, err := t.wrapped.Info(ctx)
	if err != nil {
		return nil, err
	}
	// 返回带有别名的工具信息
	return &schema.ToolInfo{
		Name:        t.alias,
		Desc:        info.Desc,
		ParamsOneOf: info.ParamsOneOf,
	}, nil
}

// InvokableRun 执行工具（委托给原工具）.
func (t *AliasedTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	return t.wrapped.InvokableRun(ctx, argumentsInJSON, opts...)
}
