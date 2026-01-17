// Package metric 提供评估指标计算功能
// 复用自 WeKnora，包含 BLEU/ROUGE/MRR/MAP/NDCG/Precision/Recall 等标准 NLP 评估指标
package metric

import (
	"github.com/yanyiwu/gojieba"
)

// Jieba 全局中文分词器实例
var Jieba *gojieba.Jieba = gojieba.NewJieba()

// MetricInput 评估指标输入数据
type MetricInput struct {
	// 检索评估相关
	RetrievalGT  [][]int // 检索 Ground Truth (查询ID -> 相关文档ID列表)
	RetrievalIDs []int   // 检索结果ID列表

	// 生成评估相关
	GeneratedTexts string // 生成的文本
	GeneratedGT    string // Ground Truth 文本
}

// RetrievalMetrics 检索评估指标结果
type RetrievalMetrics struct {
	Precision float64 `json:"precision"` // 精确率
	Recall    float64 `json:"recall"`    // 召回率
	NDCG3     float64 `json:"ndcg3"`     // NDCG@3
	NDCG10    float64 `json:"ndcg10"`    // NDCG@10
	MRR       float64 `json:"mrr"`       // 平均倒数排名
	MAP       float64 `json:"map"`       // 平均精度均值
}

// GenerationMetrics 生成评估指标结果
type GenerationMetrics struct {
	BLEU1  float64 `json:"bleu1"`  // BLEU-1
	BLEU2  float64 `json:"bleu2"`  // BLEU-2
	BLEU4  float64 `json:"bleu4"`  // BLEU-4
	ROUGE1 float64 `json:"rouge1"` // ROUGE-1
	ROUGE2 float64 `json:"rouge2"` // ROUGE-2
	ROUGEL float64 `json:"rougel"` // ROUGE-L
}

// MetricResult 完整评估结果
type MetricResult struct {
	RetrievalMetrics  RetrievalMetrics  `json:"retrieval_metrics"`
	GenerationMetrics GenerationMetrics `json:"generation_metrics"`
}

// Metric 评估指标计算接口
type Metric interface {
	Compute(input *MetricInput) float64
}
