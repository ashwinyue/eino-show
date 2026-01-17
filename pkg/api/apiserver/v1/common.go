// Package apiserverv1 提供 API 请求/响应类型定义（对齐 WeKnora）.
package apiserverv1

// ===== 通用请求/响应类型 =====

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// DeleteResponse 删除响应
type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IDRequest ID 请求（用于 URI 参数）
type IDRequest struct {
	ID string `uri:"id" binding:"required"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
