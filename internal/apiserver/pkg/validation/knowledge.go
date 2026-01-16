package validation

import (
	"context"

	"github.com/ashwinyue/eino-show/internal/pkg/errno"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ValidateListKnowledgeBases 校验获取知识库列表请求.
func (v *Validator) ValidateListKnowledgeBases(ctx context.Context, rq *v1.ListKnowledgeBasesRequest) error {
	_ = ctx
	_ = rq
	return nil
}

// ValidateGetKnowledgeBase 校验获取知识库请求.
func (v *Validator) ValidateGetKnowledgeBase(ctx context.Context, rq *v1.GetKnowledgeBaseRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("knowledge base id cannot be empty")
	}
	return nil
}

// ValidateCreateKnowledgeBase 校验创建知识库请求.
func (v *Validator) ValidateCreateKnowledgeBase(ctx context.Context, rq *v1.CreateKnowledgeBaseRequest) error {
	if rq.Name == "" {
		return errno.ErrInvalidArgument.WithMessage("name cannot be empty")
	}
	if len(rq.Name) > 255 {
		return errno.ErrInvalidArgument.WithMessage("name must be less than 255 characters")
	}
	return nil
}

// ValidateUpdateKnowledgeBase 校验更新知识库请求.
func (v *Validator) ValidateUpdateKnowledgeBase(ctx context.Context, rq *v1.UpdateKnowledgeBaseRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("knowledge base id cannot be empty")
	}
	if rq.Name != nil && len(*rq.Name) > 255 {
		return errno.ErrInvalidArgument.WithMessage("name must be less than 255 characters")
	}
	return nil
}

// ValidateDeleteKnowledgeBase 校验删除知识库请求.
func (v *Validator) ValidateDeleteKnowledgeBase(ctx context.Context, rq *v1.DeleteKnowledgeBaseRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("knowledge base id cannot be empty")
	}
	return nil
}

// ValidateGetKnowledgeStats 校验获取知识库统计请求.
func (v *Validator) ValidateGetKnowledgeStats(ctx context.Context, rq *v1.GetKnowledgeStatsRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("knowledge base id cannot be empty")
	}
	return nil
}

// ValidateListKnowledges 校验获取知识列表请求.
func (v *Validator) ValidateListKnowledges(ctx context.Context, rq *v1.ListKnowledgesRequest) error {
	if rq.KbId == "" {
		return errno.ErrInvalidArgument.WithMessage("knowledge base id cannot be empty")
	}
	return nil
}

// ValidateDeleteKnowledge 校验删除知识请求.
func (v *Validator) ValidateDeleteKnowledge(ctx context.Context, rq *v1.DeleteKnowledgeRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("knowledge id cannot be empty")
	}
	return nil
}
