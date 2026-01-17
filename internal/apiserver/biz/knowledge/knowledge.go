package knowledge

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ashwinyue/eino-show/internal/apiserver/model"
	"github.com/ashwinyue/eino-show/internal/apiserver/store"
	"github.com/ashwinyue/eino-show/internal/pkg/contextx"
	v1 "github.com/ashwinyue/eino-show/pkg/api/apiserver/v1"
	"github.com/ashwinyue/eino-show/pkg/store/where"
)

type KnowledgeBiz interface {
	CreateKB(ctx context.Context, req *v1.CreateKnowledgeBaseRequest) (*v1.CreateKnowledgeBaseResponse, error)
	GetKB(ctx context.Context, req *v1.GetKnowledgeBaseRequest) (*v1.GetKnowledgeBaseResponse, error)
	ListKB(ctx context.Context, req *v1.ListKnowledgeBasesRequest) (*v1.ListKnowledgeBasesResponse, error)
	UpdateKB(ctx context.Context, id string, req *v1.UpdateKnowledgeBaseRequest) (*v1.UpdateKnowledgeBaseResponse, error)
	DeleteKB(ctx context.Context, req *v1.DeleteKnowledgeBaseRequest) (*v1.DeleteKnowledgeBaseResponse, error)
	GetKBStats(ctx context.Context, kbID string) (*v1.GetKnowledgeStatsResponse, error)
	ListKnowledges(ctx context.Context, kbID string, req *v1.ListKnowledgesRequest) (*v1.ListKnowledgesResponse, error)
	DeleteKnowledge(ctx context.Context, req *v1.DeleteKnowledgeRequest) (*v1.DeleteKnowledgeResponse, error)
	ListChunks(ctx context.Context, knowledgeID string, req *v1.ListChunksRequest) (*v1.ListChunksResponse, error)
	GetChunk(ctx context.Context, req *v1.GetChunkRequest) (*v1.GetChunkResponse, error)
	UpdateChunk(ctx context.Context, id string, req *v1.UpdateChunkRequest) (*v1.UpdateChunkResponse, error)
	DeleteChunk(ctx context.Context, req *v1.DeleteChunkRequest) (*v1.DeleteChunkResponse, error)
	DeleteChunksByKnowledgeID(ctx context.Context, knowledgeID string) error
	HybridSearch(ctx context.Context, req *v1.HybridSearchRequest) (*v1.HybridSearchResponse, error)
}

type knowledgeBiz struct {
	store store.IStore
}

func New(store store.IStore) KnowledgeBiz {
	return &knowledgeBiz{store: store}
}

func (b *knowledgeBiz) CreateKB(ctx context.Context, req *v1.CreateKnowledgeBaseRequest) (*v1.CreateKnowledgeBaseResponse, error) {
	tenantID := contextx.TenantID(ctx)
	now := time.Now()

	kbM := &model.KnowledgeBaseM{
		ID:          uuid.New().String(),
		TenantID:    int32(tenantID),
		Name:        req.Name,
		Description: &req.Description,
		Type:        "document",
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := b.store.KnowledgeBase().Create(ctx, kbM); err != nil {
		return nil, err
	}

	return &v1.CreateKnowledgeBaseResponse{
		KnowledgeBase: toKnowledgeBaseResponse(kbM),
	}, nil
}

func (b *knowledgeBiz) GetKB(ctx context.Context, req *v1.GetKnowledgeBaseRequest) (*v1.GetKnowledgeBaseResponse, error) {
	kbM, err := b.store.KnowledgeBase().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetKnowledgeBaseResponse{
		KnowledgeBase: toKnowledgeBaseResponse(kbM),
	}, nil
}

func (b *knowledgeBiz) ListKB(ctx context.Context, req *v1.ListKnowledgeBasesRequest) (*v1.ListKnowledgeBasesResponse, error) {
	tenantID := contextx.TenantID(ctx)
	opts := where.NewWhere().F("tenant_id", tenantID)
	if req.PageSize > 0 {
		opts.P(req.Page, req.PageSize)
	}

	total, list, err := b.store.KnowledgeBase().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	kbs := make([]*v1.KnowledgeBaseResponse, len(list))
	for i, kb := range list {
		kbs[i] = toKnowledgeBaseResponse(kb)
	}

	return &v1.ListKnowledgeBasesResponse{
		KnowledgeBases: kbs,
		Total:          total,
	}, nil
}

func (b *knowledgeBiz) UpdateKB(ctx context.Context, id string, req *v1.UpdateKnowledgeBaseRequest) (*v1.UpdateKnowledgeBaseResponse, error) {
	kbM, err := b.store.KnowledgeBase().Get(ctx, where.F("id", id))
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		kbM.Name = *req.Name
	}
	if req.Description != nil {
		kbM.Description = req.Description
	}
	now := time.Now()
	kbM.UpdatedAt = &now

	if err := b.store.KnowledgeBase().Update(ctx, kbM); err != nil {
		return nil, err
	}

	return &v1.UpdateKnowledgeBaseResponse{
		KnowledgeBase: toKnowledgeBaseResponse(kbM),
	}, nil
}

func (b *knowledgeBiz) DeleteKB(ctx context.Context, req *v1.DeleteKnowledgeBaseRequest) (*v1.DeleteKnowledgeBaseResponse, error) {
	if err := b.store.KnowledgeBase().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteKnowledgeBaseResponse{
		Success: true,
	}, nil
}

func (b *knowledgeBiz) GetKBStats(ctx context.Context, kbID string) (*v1.GetKnowledgeStatsResponse, error) {
	// 统计知识和分块数量
	_, knowledges, _ := b.store.Knowledge().List(ctx, where.F("knowledge_base_id", kbID))
	_, chunks, _ := b.store.Chunk().List(ctx, where.F("knowledge_base_id", kbID))
	return &v1.GetKnowledgeStatsResponse{
		KnowledgeCount: int64(len(knowledges)),
		ChunkCount:     int64(len(chunks)),
	}, nil
}

func (b *knowledgeBiz) ListKnowledges(ctx context.Context, kbID string, req *v1.ListKnowledgesRequest) (*v1.ListKnowledgesResponse, error) {
	opts := where.NewWhere().F("knowledge_base_id", kbID)

	total, list, err := b.store.Knowledge().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	knowledges := make([]*v1.KnowledgeResponse, len(list))
	for i, k := range list {
		knowledges[i] = toKnowledgeResponse(k)
	}

	return &v1.ListKnowledgesResponse{
		Knowledge: knowledges,
		Total:     total,
	}, nil
}

func (b *knowledgeBiz) DeleteKnowledge(ctx context.Context, req *v1.DeleteKnowledgeRequest) (*v1.DeleteKnowledgeResponse, error) {
	if err := b.store.Knowledge().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteKnowledgeResponse{
		Success: true,
	}, nil
}

func (b *knowledgeBiz) ListChunks(ctx context.Context, knowledgeID string, req *v1.ListChunksRequest) (*v1.ListChunksResponse, error) {
	opts := where.NewWhere().F("knowledge_id", knowledgeID)

	total, list, err := b.store.Chunk().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	chunks := make([]*v1.ChunkResponse, len(list))
	for i, c := range list {
		chunks[i] = toChunkResponse(c)
	}

	return &v1.ListChunksResponse{
		Chunks: chunks,
		Total:  total,
	}, nil
}

func (b *knowledgeBiz) GetChunk(ctx context.Context, req *v1.GetChunkRequest) (*v1.GetChunkResponse, error) {
	chunkM, err := b.store.Chunk().Get(ctx, where.F("id", req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.GetChunkResponse{
		Chunk: toChunkResponse(chunkM),
	}, nil
}

func (b *knowledgeBiz) UpdateChunk(ctx context.Context, id string, req *v1.UpdateChunkRequest) (*v1.UpdateChunkResponse, error) {
	chunkM, err := b.store.Chunk().Get(ctx, where.F("id", id))
	if err != nil {
		return nil, err
	}

	if req.Content != "" {
		chunkM.Content = req.Content
	}

	if err := b.store.Chunk().Update(ctx, chunkM); err != nil {
		return nil, err
	}

	return &v1.UpdateChunkResponse{
		Chunk: toChunkResponse(chunkM),
	}, nil
}

func (b *knowledgeBiz) DeleteChunk(ctx context.Context, req *v1.DeleteChunkRequest) (*v1.DeleteChunkResponse, error) {
	if err := b.store.Chunk().Delete(ctx, where.F("id", req.Id)); err != nil {
		return nil, err
	}

	return &v1.DeleteChunkResponse{
		Success: true,
	}, nil
}

// DeleteChunksByKnowledgeID 删除指定知识下的所有分块.
func (b *knowledgeBiz) DeleteChunksByKnowledgeID(ctx context.Context, knowledgeID string) error {
	return b.store.Chunk().DeleteByKnowledgeID(ctx, knowledgeID)
}

func (b *knowledgeBiz) HybridSearch(ctx context.Context, req *v1.HybridSearchRequest) (*v1.HybridSearchResponse, error) {
	kbID := req.KnowledgeBaseId
	if kbID == "" {
		return &v1.HybridSearchResponse{Results: []*v1.SearchResultItem{}}, nil
	}

	// 获取知识库下的所有分块
	_, chunks, err := b.store.Chunk().List(ctx, where.F("knowledge_base_id", kbID))
	if err != nil {
		return nil, err
	}

	query := req.QueryText
	matchCount := int(req.MatchCount)
	if matchCount <= 0 {
		matchCount = 10
	}

	// RRF 融合配置
	rrfK := 60            // RRF 参数 k
	vectorWeight := 0.7   // 向量检索权重
	keywordsWeight := 0.3 // 关键词检索权重

	// 1. 关键词检索（TF-IDF 评分）
	var keywordRanked []rankedItem
	for _, chunk := range chunks {
		score := calculateRelevanceScore(chunk.Content, query)
		if score > 0 {
			keywordRanked = append(keywordRanked, rankedItem{chunk: chunk, score: score})
		}
	}
	// 按分数排序
	sortRankedItems(keywordRanked)

	// 2. 向量检索（使用 TF-IDF 模拟，后续可替换为真实向量检索）
	var vectorRanked []rankedItem
	for _, chunk := range chunks {
		score := calculateSemanticScore(chunk.Content, query)
		if score > 0 {
			vectorRanked = append(vectorRanked, rankedItem{chunk: chunk, score: score})
		}
	}
	sortRankedItems(vectorRanked)

	// 3. RRF (Reciprocal Rank Fusion) 融合
	rrfScores := make(map[string]float64)
	chunkMap := make(map[string]*model.ChunkM)

	// 处理向量检索结果
	for rank, item := range vectorRanked {
		rrfScore := vectorWeight * (1.0 / float64(rrfK+rank+1))
		rrfScores[item.chunk.ID] += rrfScore
		chunkMap[item.chunk.ID] = item.chunk
	}

	// 处理关键词检索结果
	for rank, item := range keywordRanked {
		rrfScore := keywordsWeight * (1.0 / float64(rrfK+rank+1))
		rrfScores[item.chunk.ID] += rrfScore
		chunkMap[item.chunk.ID] = item.chunk
	}

	// 4. 按 RRF 分数排序
	type fusedResult struct {
		chunkID string
		score   float64
	}
	var fused []fusedResult
	for id, score := range rrfScores {
		fused = append(fused, fusedResult{chunkID: id, score: score})
	}
	for i := 0; i < len(fused)-1; i++ {
		for j := i + 1; j < len(fused); j++ {
			if fused[j].score > fused[i].score {
				fused[i], fused[j] = fused[j], fused[i]
			}
		}
	}

	// 5. 取前 N 个结果
	var results []*v1.SearchResultItem
	for i := 0; i < len(fused) && i < matchCount; i++ {
		chunk := chunkMap[fused[i].chunkID]
		results = append(results, &v1.SearchResultItem{
			ID:          chunk.ID,
			KnowledgeID: chunk.KnowledgeID,
			Content:     chunk.Content,
			Score:       fused[i].score,
		})
	}

	return &v1.HybridSearchResponse{
		Results: results,
		Total:   int64(len(results)),
	}, nil
}

// rankedItem 排序结果项
type rankedItem struct {
	chunk *model.ChunkM
	score float64
}

// sortRankedItems 排序
func sortRankedItems(items []rankedItem) {
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].score > items[i].score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// calculateSemanticScore 计算语义相似度（简化版，后续可替换为向量相似度）
func calculateSemanticScore(content, query string) float64 {
	// 使用 n-gram 重叠作为语义相似度的近似
	if query == "" || content == "" {
		return 0
	}

	queryLower := toLower(query)
	contentLower := toLower(content)

	// 计算 bigram 重叠
	queryBigrams := getBigrams(queryLower)
	contentBigrams := getBigrams(contentLower)

	if len(queryBigrams) == 0 {
		return calculateRelevanceScore(content, query)
	}

	overlap := 0
	for bg := range queryBigrams {
		if contentBigrams[bg] {
			overlap++
		}
	}

	return float64(overlap) / float64(len(queryBigrams))
}

// getBigrams 获取字符串的 bigrams
func getBigrams(s string) map[string]bool {
	bigrams := make(map[string]bool)
	runes := []rune(s)
	for i := 0; i < len(runes)-1; i++ {
		bigrams[string(runes[i:i+2])] = true
	}
	return bigrams
}

// calculateRelevanceScore 计算相关性分数
func calculateRelevanceScore(content, query string) float64 {
	if query == "" {
		return 1.0
	}
	if content == "" {
		return 0
	}

	// 统计查询词出现次数
	queryLower := toLower(query)
	contentLower := toLower(content)

	count := 0
	pos := 0
	for {
		idx := findSubstringFrom(contentLower, queryLower, pos)
		if idx < 0 {
			break
		}
		count++
		pos = idx + 1
	}

	if count == 0 {
		return 0
	}

	// 基于词频和内容长度计算分数
	tf := float64(count) / float64(len(content)/100+1)
	return tf * 10 // 归一化分数
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func findSubstringFrom(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// containsQuery 简单的查询包含检查.
func containsQuery(content, query string) bool {
	return len(query) > 0 && len(content) > 0 &&
		(len(content) >= len(query) && (content == query ||
			findSubstring(content, query)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toKnowledgeBaseResponse(kb *model.KnowledgeBaseM) *v1.KnowledgeBaseResponse {
	resp := &v1.KnowledgeBaseResponse{
		ID:       kb.ID,
		Name:     kb.Name,
		TenantID: uint64(kb.TenantID),
		Type:     kb.Type,
	}
	if kb.Description != nil {
		resp.Description = *kb.Description
	}
	if kb.CreatedAt != nil {
		resp.CreatedAt = *kb.CreatedAt
	}
	if kb.UpdatedAt != nil {
		resp.UpdatedAt = *kb.UpdatedAt
	}
	return resp
}

func toKnowledgeResponse(k *model.KnowledgeM) *v1.KnowledgeResponse {
	resp := &v1.KnowledgeResponse{
		ID:     k.ID,
		KbID:   k.KnowledgeBaseID,
		Title:  k.Title,
		Type:   k.Type,
		Status: k.ParseStatus,
	}
	if k.FileSize != nil {
		resp.FileSize = *k.FileSize
	}
	if k.CreatedAt != nil {
		resp.CreatedAt = *k.CreatedAt
	}
	if k.UpdatedAt != nil {
		resp.UpdatedAt = *k.UpdatedAt
	}
	return resp
}

func toChunkResponse(c *model.ChunkM) *v1.ChunkResponse {
	resp := &v1.ChunkResponse{
		ID:          c.ID,
		KnowledgeID: c.KnowledgeID,
		Content:     c.Content,
		Index:       int(c.ChunkIndex),
	}
	if c.CreatedAt != nil {
		resp.CreatedAt = *c.CreatedAt
	}
	return resp
}
