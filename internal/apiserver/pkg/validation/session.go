package validation

import (
	"context"

	"github.com/ashwinyue/eino-show/internal/pkg/errno"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// ValidateCreateSession 校验创建会话请求.
func (v *Validator) ValidateCreateSession(ctx context.Context, rq *v1.CreateSessionRequest) error {
	if rq.Title == "" {
		return errno.ErrInvalidArgument.WithMessage("title cannot be empty")
	}
	if len(rq.Title) > 255 {
		return errno.ErrInvalidArgument.WithMessage("title must be less than 255 characters")
	}
	return nil
}

// ValidateGetSession 校验获取会话请求.
func (v *Validator) ValidateGetSession(ctx context.Context, rq *v1.GetSessionRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("session id cannot be empty")
	}
	return nil
}

// ValidateListSessions 校验获取会话列表请求.
func (v *Validator) ValidateListSessions(ctx context.Context, rq *v1.ListSessionsRequest) error {
	return nil
}

// ValidateUpdateSession 校验更新会话请求.
func (v *Validator) ValidateUpdateSession(ctx context.Context, rq *v1.UpdateSessionRequest) error {
	// UpdateSessionRequest 不包含 Id 字段，由路由参数提供
	if rq.Title != nil && len(*rq.Title) > 255 {
		return errno.ErrInvalidArgument.WithMessage("title must be less than 255 characters")
	}
	return nil
}

// ValidateDeleteSession 校验删除会话请求.
func (v *Validator) ValidateDeleteSession(ctx context.Context, rq *v1.DeleteSessionRequest) error {
	if rq.Id == "" {
		return errno.ErrInvalidArgument.WithMessage("session id cannot be empty")
	}
	return nil
}

// ValidateStopSession 校验停止会话请求.
func (v *Validator) ValidateStopSession(ctx context.Context, rq *v1.StopSessionRequest) error {
	if rq.MessageID == "" {
		return errno.ErrInvalidArgument.WithMessage("message id cannot be empty")
	}
	return nil
}
