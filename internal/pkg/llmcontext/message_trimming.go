// Package llmcontext provides message trimming utilities.
// Reference: WeKnora llmcontext/message_trimming.go
package llmcontext

import (
	"github.com/cloudwego/eino/schema"
)

// TrimMessagesWithConsistency 确保消息一致性.
// 删除消息时，同时删除其关联的 Tool 结果，避免孤立消息.
//
// 规则：
// 1. 系统消息永久保留
// 2. 删除 Assistant 消息时，同时删除其调用的所有 Tool 结果
// 3. 删除 Tool 结果时，同时删除其对应的 Assistant 消息
// 4. 保留消息的完整性（Assistant + Tool 结果成对出现）
func TrimMessagesWithConsistency(messages []*schema.Message, maxItems int) []*schema.Message {
	if len(messages) <= maxItems {
		return messages
	}

	// 构建消息索引和关系图
	msgGraph := buildMessageGraph(messages)

	// 标记需要保留的消息
	keepCount := 0
	keptIndices := make(map[int]bool)

	// 1. 首先保留所有系统消息
	for i, msg := range messages {
		if msg.Role == schema.System {
			keptIndices[i] = true
			keepCount++
		}
	}

	// 2. 从最新到最旧保留对话对（倒序遍历）
	for i := len(messages) - 1; i >= 0; i-- {
		if keepCount >= maxItems {
			break
		}

		// 跳过系统消息和已保留的消息
		if messages[i].Role == schema.System || keptIndices[i] {
			continue
		}

		// 检查是否是 Assistant 消息且有 Tool Call
		if messages[i].Role == schema.Assistant && len(messages[i].ToolCalls) > 0 {
			keptIndices[i] = true
			keepCount++

			// 保留所有关联的 Tool 结果
			for _, toolCall := range messages[i].ToolCalls {
				if toolResultIdx, ok := msgGraph.toolCallToResult[toolCall.ID]; ok {
					if !keptIndices[toolResultIdx] {
						keptIndices[toolResultIdx] = true
						keepCount++
					}
				}
			}
		} else if messages[i].Role == schema.Tool {
			// 检查这个 Tool 结果的 Assistant 消息是否已被保留
			if assistantIdx, ok := msgGraph.toolToAssistant[i]; ok {
				if !keptIndices[assistantIdx] && !keptIndices[i] {
					keptIndices[assistantIdx] = true
					keepCount++
					keptIndices[i] = true
					keepCount++
				}
			} else {
				// 孤立的 Tool 消息，直接保留
				if !keptIndices[i] {
					keptIndices[i] = true
					keepCount++
				}
			}
		} else {
			// 普通消息（user, assistant without tools）
			if !keptIndices[i] {
				keptIndices[i] = true
				keepCount++
			}
		}
	}

	// 构建结果
	result := make([]*schema.Message, 0, keepCount)
	for i := 0; i < len(messages); i++ {
		if keptIndices[i] {
			result = append(result, messages[i])
		}
	}

	return result
}

// RemoveOrphanedToolMessages 移除孤立的 Tool 消息.
// 孤立的 Tool 消息是指其对应的 Assistant 消息已被删除的消息.
func RemoveOrphanedToolMessages(messages []*schema.Message) []*schema.Message {
	graph := buildMessageGraph(messages)
	hasValidAssistant := make(map[int]bool)

	// 标记所有有有效 Assistant 消息的 Tool 结果
	for toolIdx, assistantIdx := range graph.toolToAssistant {
		if assistantIdx >= 0 && assistantIdx < len(messages) {
			hasValidAssistant[toolIdx] = true
		}
	}

	// 构建结果，只保留有有效 Assistant 的 Tool 消息
	result := make([]*schema.Message, 0, len(messages))
	for i, msg := range messages {
		if msg.Role != schema.Tool || hasValidAssistant[i] {
			result = append(result, msg)
		}
	}

	return result
}

// messageGraph 消息关系图.
type messageGraph struct {
	toolCallToResult   map[string]int   // Tool Call ID -> Tool 结果消息索引
	toolToAssistant    map[int]int      // Tool 消息索引 -> Assistant 消息索引
	assistantToolCalls map[int][]string // Assistant 消息索引 -> Tool Call IDs
}

// buildMessageGraph 构建消息关系图.
func buildMessageGraph(messages []*schema.Message) messageGraph {
	graph := messageGraph{
		toolCallToResult:   make(map[string]int),
		toolToAssistant:    make(map[int]int),
		assistantToolCalls: make(map[int][]string),
	}

	// 第一遍：收集所有 Assistant 消息的 Tool Call
	for i, msg := range messages {
		if msg.Role == schema.Assistant && len(msg.ToolCalls) > 0 {
			toolCallIDs := make([]string, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				toolCallIDs = append(toolCallIDs, tc.ID)
			}
			graph.assistantToolCalls[i] = toolCallIDs
		}
	}

	// 第二遍：建立 Tool 结果到 Assistant 的映射
	for i, msg := range messages {
		if msg.Role == schema.Tool && msg.ToolCallID != "" {
			for assistantIdx, toolCallIDs := range graph.assistantToolCalls {
				for _, tcID := range toolCallIDs {
					if tcID == msg.ToolCallID {
						graph.toolCallToResult[tcID] = i
						graph.toolToAssistant[i] = assistantIdx
						break
					}
				}
			}
		}
	}

	return graph
}

// ValidateMessageConsistency 验证消息一致性.
// 返回孤立 Tool 消息的数量.
func ValidateMessageConsistency(messages []*schema.Message) int {
	graph := buildMessageGraph(messages)
	orphanCount := 0

	for i, msg := range messages {
		if msg.Role == schema.Tool {
			if assistantIdx, ok := graph.toolToAssistant[i]; !ok || assistantIdx < 0 || assistantIdx >= len(messages) {
				orphanCount++
			}
		}
	}

	return orphanCount
}

// GetToolCallPairs 获取所有 Tool Call 对.
// 返回 Assistant 消息索引 -> Tool 结果消息索引列表 的映射.
func GetToolCallPairs(messages []*schema.Message) map[int][]int {
	graph := buildMessageGraph(messages)
	pairs := make(map[int][]int)

	for toolIdx, assistantIdx := range graph.toolToAssistant {
		pairs[assistantIdx] = append(pairs[assistantIdx], toolIdx)
	}

	return pairs
}

// TrimToTokenLimit 将消息裁剪到指定的 token 限制.
func TrimToTokenLimit(messages []*schema.Message, maxTokens int, estimator func([]*schema.Message) int) []*schema.Message {
	if estimator == nil {
		// 默认估算器：4 字符 ≈ 1 token
		estimator = func(msgs []*schema.Message) int {
			total := 0
			for _, msg := range msgs {
				total += len(msg.Content) / 4
			}
			return total
		}
	}

	currentTokens := estimator(messages)
	if currentTokens <= maxTokens {
		return messages
	}

	// 逐步减少消息数量
	for maxItems := len(messages) - 1; maxItems > 1; maxItems-- {
		trimmed := TrimMessagesWithConsistency(messages, maxItems)
		if estimator(trimmed) <= maxTokens {
			return RemoveOrphanedToolMessages(trimmed)
		}
	}

	// 最少保留系统消息
	result := make([]*schema.Message, 0)
	for _, msg := range messages {
		if msg.Role == schema.System {
			result = append(result, msg)
		}
	}
	return result
}
