package validation

import (
	"context"

	genericvalidation "github.com/ashwinyue/eino-show/pkg/validation"

	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	"github.com/ashwinyue/eino-show/internal/pkg/errno"
	apiv1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

func (v *Validator) ValidateUserRules() genericvalidation.Rules {
	// 通用的密码校验函数
	validatePassword := func() genericvalidation.ValidatorFunc {
		return func(value any) error {
			return isValidPassword(value.(string))
		}
	}

	// 定义各字段的校验逻辑，通过一个 map 实现模块化和简化
	return genericvalidation.Rules{
		"Password":    validatePassword(),
		"OldPassword": validatePassword(),
		"NewPassword": validatePassword(),
		"UserID": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("userID cannot be empty")
			}
			return nil
		},
		"Username": func(value any) error {
			if !isValidUsername(value.(string)) {
				return errno.ErrUsernameInvalid
			}
			return nil
		},
		"Nickname": func(value any) error {
			if len(value.(string)) >= 30 {
				return errno.ErrInvalidArgument.WithMessage("nickname must be less than 30 characters")
			}
			return nil
		},
		"Email": func(value any) error {
			return isValidEmail(value.(string))
		},
		"Phone": func(value any) error {
			return isValidPhone(value.(string))
		},
		"Limit": func(value any) error {
			if value.(int64) <= 0 {
				return errno.ErrInvalidArgument.WithMessage("limit must be greater than 0")
			}
			return nil
		},
		"Offset": func(value any) error {
			return nil
		},
	}
}

// ValidateLoginRequest 校验登录请求（对齐 WeKnora，使用 email 登录）.
func (v *Validator) ValidateLoginRequest(ctx context.Context, rq *apiv1.LoginRequest) error {
	// email 必须提供
	if rq.Email == "" {
		return errno.ErrInvalidArgument.WithMessage("email is required")
	}

	// password 必须提供
	if rq.Password == "" {
		return errno.ErrInvalidArgument.WithMessage("password is required")
	}

	// 验证 email 格式
	if err := isValidEmail(rq.Email); err != nil {
		return err
	}

	return nil
}

// ValidateChangePasswordRequest 校验 ChangePasswordRequest 结构体的有效性.
func (v *Validator) ValidateChangePasswordRequest(ctx context.Context, rq *apiv1.ChangePasswordRequest) error {
	if rq.UserID != contextx.UserID(ctx) {
		return errno.ErrPermissionDenied.WithMessage("The logged-in user `%s` does not match request user `%s`", contextx.UserID(ctx), rq.UserID)
	}
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateCreateUserRequest 校验 CreateUserRequest 结构体的有效性.
func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *apiv1.CreateUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateUpdateUserRequest 校验更新用户请求.
func (v *Validator) ValidateUpdateUserRequest(ctx context.Context, rq *apiv1.UpdateUserRequest) error {
	if rq.UserID != contextx.UserID(ctx) {
		return errno.ErrPermissionDenied.WithMessage("The logged-in user `%s` does not match request user `%s`", contextx.UserID(ctx), rq.UserID)
	}
	return genericvalidation.ValidateSelectedFields(rq, v.ValidateUserRules(), "UserID")
}

// ValidateDeleteUserRequest 校验 DeleteUserRequest 结构体的有效性.
func (v *Validator) ValidateDeleteUserRequest(ctx context.Context, rq *apiv1.DeleteUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateGetUserRequest 校验 GetUserRequest 结构体的有效性.
func (v *Validator) ValidateGetUserRequest(ctx context.Context, rq *apiv1.GetUserRequest) error {
	if rq.UserID != contextx.UserID(ctx) {
		return errno.ErrPermissionDenied.WithMessage("The logged-in user `%s` does not match request user `%s`", contextx.UserID(ctx), rq.UserID)
	}
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

// ValidateListUserRequest 校验 ListUsersRequest 结构体的有效性.
func (v *Validator) ValidateListUserRequest(ctx context.Context, rq *apiv1.ListUsersRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}
