<template>
  <div class="agent-transfer-display">
    <div class="transfer-header">
      <span class="transfer-icon">🔄</span>
      <span class="transfer-title">任务转移完成</span>
    </div>
    <div class="transfer-content">
      <div v-if="data.agent_id" class="transfer-target">
        <span class="label">目标 Agent:</span>
        <span class="value">{{ data.agent_name || formatAgentId(data.agent_id) }}</span>
      </div>
      <div v-if="data.duration_ms" class="transfer-duration">
        <span class="label">执行时长:</span>
        <span class="value">{{ formatDuration(data.duration_ms) }}</span>
      </div>
      <div v-if="data.answer" class="transfer-answer">
        <div class="answer-label">Agent 回答:</div>
        <div class="answer-text markdown-content" v-html="renderMarkdown(data.answer)"></div>
      </div>
      <div v-if="data.error" class="transfer-error">
        <span class="error-icon">⚠️</span>
        <span class="error-text">{{ data.error }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { AgentTransferData } from '@/types/tool-results';
import DOMPurify from 'dompurify';
import { marked } from 'marked';

interface Props {
  data: AgentTransferData;
}

const props = defineProps<Props>();

const formatAgentId = (id: string): string => {
  if (id.startsWith('builtin-')) {
    const name = id.replace('builtin-', '').replace(/-/g, ' ');
    return name.charAt(0).toUpperCase() + name.slice(1);
  }
  return id;
};

const formatDuration = (ms: number): string => {
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
};

const renderMarkdown = (content: string): string => {
  try {
    const html = marked.parse(content) as string;
    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'u', 'code', 'pre', 'ul', 'ol', 'li', 'blockquote', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'a', 'span'],
      ALLOWED_ATTR: ['href', 'title', 'target', 'rel', 'class']
    });
  } catch {
    return content;
  }
};
</script>

<style lang="less" scoped>
.agent-transfer-display {
  padding: 12px 16px;
  background: linear-gradient(135deg, rgba(7, 192, 95, 0.05), rgba(7, 192, 95, 0.02));
  border-radius: 6px;
  border: 1px solid rgba(7, 192, 95, 0.2);
}

.transfer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(7, 192, 95, 0.1);
}

.transfer-icon {
  font-size: 16px;
}

.transfer-title {
  font-size: 13px;
  font-weight: 600;
  color: #07c05f;
}

.transfer-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.transfer-target,
.transfer-duration {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.label {
  color: #6b7280;
  font-weight: 500;
}

.value {
  color: #374151;
  font-weight: 500;
}

.transfer-answer {
  margin-top: 4px;
}

.answer-label {
  font-size: 12px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 6px;
}

.answer-text {
  padding: 8px 12px;
  background: #ffffff;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
  font-size: 13px;
  line-height: 1.6;
  color: #374151;

  :deep(p) {
    margin: 4px 0;
  }

  :deep(code) {
    background: #f3f4f6;
    padding: 2px 5px;
    border-radius: 3px;
    font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
    font-size: 11px;
  }
}

.transfer-error {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: rgba(227, 77, 89, 0.1);
  border-radius: 4px;
  font-size: 12px;
  color: #e34d59;
}

.error-icon {
  font-size: 14px;
}
</style>
