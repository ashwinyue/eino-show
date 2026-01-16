// Copyright 2026 阿斯温月 <stary99c@163.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/ashwinyue/eino-show. The professional
// version of this repository is https://github.com/onexstack/onex.

// Package tool 提供 Eino 工具的注册和管理.
package tool

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

// InvokableTool 工具接口别名.
type InvokableTool = tool.InvokableTool

// Registry 工具注册表，管理所有可用工具.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]InvokableTool
}

// NewRegistry 创建新的工具注册表.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]InvokableTool),
	}
}

// Register 注册一个工具.
func (r *Registry) Register(t InvokableTool) {
	if t == nil {
		return
	}
	info, _ := t.Info(context.Background())
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[info.Name] = t
}

// Unregister 注销一个工具.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

// Get 获取指定名称的工具.
func (r *Registry) Get(name string) (InvokableTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List 列出所有工具.
func (r *Registry) List() []InvokableTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]InvokableTool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// GetAllTools 获取所有工具.
func (r *Registry) GetAllTools(_ context.Context) []InvokableTool {
	return r.List()
}

// GetToolsByNames 根据名称列表获取工具.
func (r *Registry) GetToolsByNames(names []string) []InvokableTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]InvokableTool, 0, len(names))
	for _, name := range names {
		if t, ok := r.tools[name]; ok {
			result = append(result, t)
		}
	}
	return result
}

// Has 检查工具是否存在.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name]
	return ok
}

// Count 返回工具数量.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Clear 清空所有工具.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]InvokableTool)
}

// Cleanup 清理工具资源.
func (r *Registry) Cleanup(_ context.Context) {
	r.Clear()
}
