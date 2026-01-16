// Package knowledge 提供知识库业务逻辑.
package knowledge

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/onexstack/onexstack/pkg/store/where"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KnowledgeBiz 知识库业务接口.
// 使用 proto 生成的类型作为请求和响应
type KnowledgeBiz interface {
	// CreateKnowledgeBase 创建知识库
	CreateKnowledgeBase(ctx context.Context, req *v1.CreateKnowledgeBaseRequest) (*v1.CreateKnowledgeBaseResponse, error)
	// GetKnowledgeBase 获取知识库详情
	GetKnowledgeBase(ctx context.Context, req *v1.GetKnowledgeBaseRequest) (*v1.GetKnowledgeBaseResponse, error)
	// ListKnowledgeBases 获取知识库列表
	ListKnowledgeBases(ctx context.Context, req *v1.ListKnowledgeBasesRequest) (*v1.ListKnowledgeBasesResponse, error)
	// UpdateKnowledgeBase 更新知识库
	UpdateKnowledgeBase(ctx context.Context, req *v1.UpdateKnowledgeBaseRequest) (*v1.UpdateKnowledgeBaseResponse, error)
	// DeleteKnowledgeBase 删除知识库
	DeleteKnowledgeBase(ctx context.Context, req *v1.DeleteKnowledgeBaseRequest) (*v1.DeleteKnowledgeBaseResponse, error)

	// GetKnowledgeStats 获取知识库统计信息
	GetKnowledgeStats(ctx context.Context, req *v1.GetKnowledgeStatsRequest) (*v1.KnowledgeStatsResponse, error)
	// ListKnowledges 获取知识列表
	ListKnowledges(ctx context.Context, req *v1.ListKnowledgesRequest) (*v1.ListKnowledgesResponse, error)
	// DeleteKnowledge 删除知识项（及其关联的分块）
	DeleteKnowledge(ctx context.Context, req *v1.DeleteKnowledgeRequest) (*v1.DeleteKnowledgeResponse, error)
}

type knowledgeBiz struct {
	store store.IStore
}

// New 创建 KnowledgeBiz 实例.
func New(store store.IStore) KnowledgeBiz {
	return &knowledgeBiz{store: store}
}

// 确保 knowledgeBiz 实现了 KnowledgeBiz 接口.
var _ KnowledgeBiz = (*knowledgeBiz)(nil)

// CreateKnowledgeBase 创建知识库.
func (b *knowledgeBiz) CreateKnowledgeBase(ctx context.Context, req *v1.CreateKnowledgeBaseRequest) (*v1.CreateKnowledgeBaseResponse, error) {
	tenantID := contextx.TenantID(ctx)
	kbM := &model.KnowledgeBaseM{
		TenantID:              int32(tenantID),
		Name:                  req.Name,
		Description:           stringPtr(req.Description),
		ChunkingConfig:        protoChunkingConfigToJSON(req.ChunkingConfig),
		ImageProcessingConfig: protoImageConfigToJSON(req.ImageProcessingConfig),
		EmbeddingModelID:      req.EmbeddingModelId,
		SummaryModelID:        req.SummaryModelId,
	}

	if err := b.store.KnowledgeBase().Create(ctx, kbM); err != nil {
		return nil, err
	}

	return &v1.CreateKnowledgeBaseResponse{
		Id:        kbM.ID,
		Name:      kbM.Name,
		CreatedAt: timePtrToProto(kbM.CreatedAt),
	}, nil
}

// GetKnowledgeBase 获取知识库详情.
func (b *knowledgeBiz) GetKnowledgeBase(ctx context.Context, req *v1.GetKnowledgeBaseRequest) (*v1.GetKnowledgeBaseResponse, error) {
	kbM, err := b.store.KnowledgeBase().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetKnowledgeBaseResponse{
		KnowledgeBase: modelKnowledgeBaseToProto(kbM),
	}, nil
}

// ListKnowledgeBases 获取知识库列表.
func (b *knowledgeBiz) ListKnowledgeBases(ctx context.Context, req *v1.ListKnowledgeBasesRequest) (*v1.ListKnowledgeBasesResponse, error) {
	_ = req
	tenantID := contextx.TenantID(ctx)
	kbs, err := b.store.KnowledgeBase().GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	pbKbs := make([]*v1.KnowledgeBase, 0, len(kbs))
	for _, kb := range kbs {
		pbKbs = append(pbKbs, modelKnowledgeBaseToProto(kb))
	}

	return &v1.ListKnowledgeBasesResponse{
		KnowledgeBases: pbKbs,
		Total:          int64(len(kbs)),
	}, nil
}

// UpdateKnowledgeBase 更新知识库.
func (b *knowledgeBiz) UpdateKnowledgeBase(ctx context.Context, req *v1.UpdateKnowledgeBaseRequest) (*v1.UpdateKnowledgeBaseResponse, error) {
	kbM, err := b.store.KnowledgeBase().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		kbM.Name = *req.Name
	}
	if req.Description != nil {
		kbM.Description = req.Description
	}
	if req.ChunkingConfig != nil {
		kbM.ChunkingConfig = protoChunkingConfigToJSON(req.ChunkingConfig)
	}
	if req.ImageProcessingConfig != nil {
		kbM.ImageProcessingConfig = protoImageConfigToJSON(req.ImageProcessingConfig)
	}
	if req.EmbeddingModelId != nil {
		kbM.EmbeddingModelID = *req.EmbeddingModelId
	}
	if req.SummaryModelId != nil {
		kbM.SummaryModelID = *req.SummaryModelId
	}

	if err := b.store.KnowledgeBase().Update(ctx, kbM); err != nil {
		return nil, err
	}

	return &v1.UpdateKnowledgeBaseResponse{
		Id:        kbM.ID,
		Name:      kbM.Name,
		UpdatedAt: timePtrToProto(kbM.UpdatedAt),
	}, nil
}

// DeleteKnowledgeBase 删除知识库.
func (b *knowledgeBiz) DeleteKnowledgeBase(ctx context.Context, req *v1.DeleteKnowledgeBaseRequest) (*v1.DeleteKnowledgeBaseResponse, error) {
	if err := b.store.KnowledgeBase().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteKnowledgeBaseResponse{
		Success: true,
		Message: "Knowledge base deleted successfully",
	}, nil
}

// GetKnowledgeStats 获取知识库统计信息.
func (b *knowledgeBiz) GetKnowledgeStats(ctx context.Context, req *v1.GetKnowledgeStatsRequest) (*v1.KnowledgeStatsResponse, error) {
	// 获取知识项数量
	knowledges, err := b.store.Knowledge().GetByKnowledgeBaseID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// 获取分块数量（通过 knowledge 表统计）
	count := int64(len(knowledges))

	return &v1.KnowledgeStatsResponse{
		KnowledgeCount: count,
		ChunkCount:     0, // TODO: 实现 Chunk 统计
	}, nil
}

// ListKnowledges 获取知识列表.
func (b *knowledgeBiz) ListKnowledges(ctx context.Context, req *v1.ListKnowledgesRequest) (*v1.ListKnowledgesResponse, error) {
	knowledges, err := b.store.Knowledge().GetByKnowledgeBaseID(ctx, req.KbId)
	if err != nil {
		return nil, err
	}

	pbKnowledges := make([]*v1.Knowledge, 0, len(knowledges))
	for _, k := range knowledges {
		pbKnowledges = append(pbKnowledges, modelKnowledgeToProto(k))
	}

	return &v1.ListKnowledgesResponse{
		Knowledge: pbKnowledges,
		Total:     int64(len(knowledges)),
	}, nil
}

// DeleteKnowledge 删除知识项.
func (b *knowledgeBiz) DeleteKnowledge(ctx context.Context, req *v1.DeleteKnowledgeRequest) (*v1.DeleteKnowledgeResponse, error) {
	if err := b.store.Knowledge().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	// TODO: 同时删除关联的分块
	// chunks, _ := b.store.Chunk().GetByKnowledgeID(ctx, req.Id)
	// for _, chunk := range chunks {
	// 	b.store.Chunk().Delete(ctx, where.F("id", chunk.ID))
	// }

	return &v1.DeleteKnowledgeResponse{
		Success: true,
		Message: "Knowledge deleted successfully",
	}, nil
}

// ===== 类型转换函数 =====

// modelKnowledgeBaseToProto 将 Model 转换为 Proto
func modelKnowledgeBaseToProto(kb *model.KnowledgeBaseM) *v1.KnowledgeBase {
	description := ""
	if kb.Description != nil {
		description = *kb.Description
	}
	pb := &v1.KnowledgeBase{
		Id:                    kb.ID,
		Name:                  kb.Name,
		Description:           description,
		ChunkingConfig:        protoJSONToProtoChunkingConfig(kb.ChunkingConfig),
		ImageProcessingConfig: protoJSONToProtoImageConfig(kb.ImageProcessingConfig),
		EmbeddingModelId:      kb.EmbeddingModelID,
		SummaryModelId:        kb.SummaryModelID,
		CreatedAt:             timePtrToProto(kb.CreatedAt),
		UpdatedAt:             timePtrToProto(kb.UpdatedAt),
	}
	return pb
}

// modelKnowledgeToProto 将 Model Knowledge 转换为 Proto
func modelKnowledgeToProto(k *model.KnowledgeM) *v1.Knowledge {
	description := ""
	if k.Description != nil {
		description = *k.Description
	}
	embeddingModelID := ""
	if k.EmbeddingModelID != nil {
		embeddingModelID = *k.EmbeddingModelID
	}
	fileName := ""
	if k.FileName != nil {
		fileName = *k.FileName
	}
	fileType := ""
	if k.FileType != nil {
		fileType = *k.FileType
	}
	return &v1.Knowledge{
		Id:               k.ID,
		KnowledgeBaseId:  k.KnowledgeBaseID,
		Type:             k.Type,
		Title:            k.Title,
		Description:      description,
		Source:           k.Source,
		ParseStatus:      k.ParseStatus,
		EnableStatus:     k.EnableStatus,
		EmbeddingModelId: embeddingModelID,
		FileName:         fileName,
		FileType:         fileType,
		CreatedAt:        timePtrToProto(k.CreatedAt),
		UpdatedAt:        timePtrToProto(k.UpdatedAt),
	}
}

// ===== JSON 序列化/反序列化函数 =====

// protoChunkingConfigToJSON 将 Proto ChunkingConfig 转换为 JSON 字符串
func protoChunkingConfigToJSON(cfg *v1.KnowledgeChunkingConfig) string {
	if cfg == nil {
		return "{}"
	}
	data, _ := json.Marshal(cfg)
	return string(data)
}

// protoImageConfigToJSON 将 Proto ImageConfig 转换为 JSON 字符串
func protoImageConfigToJSON(cfg *v1.KnowledgeImageConfig) string {
	if cfg == nil {
		return "{}"
	}
	data, _ := json.Marshal(cfg)
	return string(data)
}

// protoJSONToProtoChunkingConfig 从 JSON 字符串解析为 Proto ChunkingConfig
func protoJSONToProtoChunkingConfig(s string) *v1.KnowledgeChunkingConfig {
	if s == "" {
		return nil
	}
	var cfg v1.KnowledgeChunkingConfig
	if err := json.Unmarshal([]byte(s), &cfg); err != nil {
		// 返回默认配置
		return &v1.KnowledgeChunkingConfig{
			ChunkSize:     512,
			ChunkOverlap:  50,
			SplitMarkers:  []string{"\n\n", "\n", "。"},
			KeepSeparator: true,
		}
	}
	return &cfg
}

// protoJSONToProtoImageConfig 从 JSON 字符串解析为 Proto ImageConfig
func protoJSONToProtoImageConfig(s string) *v1.KnowledgeImageConfig {
	if s == "" {
		return nil
	}
	var cfg v1.KnowledgeImageConfig
	if err := json.Unmarshal([]byte(s), &cfg); err != nil {
		return &v1.KnowledgeImageConfig{
			EnableMultimodal: false,
			ModelId:          "",
		}
	}
	return &cfg
}

// ===== 辅助函数 =====

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtrToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
