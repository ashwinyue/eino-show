package conversion

import (
	"time"

	"github.com/ashwinyue/eino-show/pkg/core"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	apiv1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
)

// UserModelToUserV1 将模型层的 UserM 转换为 v1 User.
func UserModelToUserV1(userModel *model.UserM) *apiv1.User {
	user := &apiv1.User{
		ID:       userModel.ID,
		Username: userModel.Username,
		Email:    userModel.Email,
	}

	// 处理可选字段
	if userModel.TenantID != nil {
		user.TenantID = uint64(*userModel.TenantID)
	}
	if userModel.CreatedAt != nil {
		user.CreatedAt = *userModel.CreatedAt
	}
	if userModel.UpdatedAt != nil {
		user.UpdatedAt = *userModel.UpdatedAt
	}

	return user
}

// UserV1ToUserModel 将 v1 User 转换为模型层的 UserM.
func UserV1ToUserModel(user *apiv1.User) *model.UserM {
	var userModel model.UserM
	_ = core.CopyWithConverters(&userModel, user)
	return &userModel
}

// TenantMToTenantV1 将模型层的 TenantM 转换为 v1 Tenant.
func TenantMToTenantV1(tenantModel *model.TenantM) *apiv1.Tenant {
	tenant := &apiv1.Tenant{
		ID:   uint64(tenantModel.ID),
		Name: tenantModel.Name,
	}
	if tenantModel.CreatedAt != nil {
		tenant.CreatedAt = tenantModel.CreatedAt.Format(time.RFC3339)
	}
	if tenantModel.UpdatedAt != nil {
		tenant.UpdatedAt = tenantModel.UpdatedAt.Format(time.RFC3339)
	}
	return tenant
}

// TimePtrToTime 将 *time.Time 转换为 time.Time.
func TimePtrToTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
