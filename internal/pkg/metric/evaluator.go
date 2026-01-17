package metric

// Evaluator 评估器，提供便捷的指标计算方法
type Evaluator struct {
	// 检索指标
	precision *PrecisionMetric
	recall    *RecallMetric
	mrr       *MRRMetric
	mapMetric *MAPMetric
	ndcg3     *NDCGMetric
	ndcg10    *NDCGMetric

	// 生成指标
	bleu1  *BLEUMetric
	bleu2  *BLEUMetric
	bleu4  *BLEUMetric
	rouge1 *RougeMetric
	rouge2 *RougeMetric
	rougeL *RougeMetric
}

// NewEvaluator 创建评估器实例
func NewEvaluator() *Evaluator {
	return &Evaluator{
		// 检索指标
		precision: NewPrecisionMetric(),
		recall:    NewRecallMetric(),
		mrr:       NewMRRMetric(),
		mapMetric: NewMAPMetric(),
		ndcg3:     NewNDCGMetric(3),
		ndcg10:    NewNDCGMetric(10),

		// 生成指标 (BLEU 使用平滑)
		bleu1:  NewBLEUMetric(true, BLEU1Gram),
		bleu2:  NewBLEUMetric(true, BLEU2Gram),
		bleu4:  NewBLEUMetric(true, BLEU4Gram),
		rouge1: NewRougeMetric(false, "rouge-1", "f"),
		rouge2: NewRougeMetric(false, "rouge-2", "f"),
		rougeL: NewRougeMetric(false, "rouge-l", "f"),
	}
}

// ComputeRetrievalMetrics 计算所有检索指标
func (e *Evaluator) ComputeRetrievalMetrics(input *MetricInput) RetrievalMetrics {
	return RetrievalMetrics{
		Precision: e.precision.Compute(input),
		Recall:    e.recall.Compute(input),
		MRR:       e.mrr.Compute(input),
		MAP:       e.mapMetric.Compute(input),
		NDCG3:     e.ndcg3.Compute(input),
		NDCG10:    e.ndcg10.Compute(input),
	}
}

// ComputeGenerationMetrics 计算所有生成指标
func (e *Evaluator) ComputeGenerationMetrics(input *MetricInput) GenerationMetrics {
	return GenerationMetrics{
		BLEU1:  e.bleu1.Compute(input),
		BLEU2:  e.bleu2.Compute(input),
		BLEU4:  e.bleu4.Compute(input),
		ROUGE1: e.rouge1.Compute(input),
		ROUGE2: e.rouge2.Compute(input),
		ROUGEL: e.rougeL.Compute(input),
	}
}

// ComputeAll 计算所有指标
func (e *Evaluator) ComputeAll(input *MetricInput) MetricResult {
	return MetricResult{
		RetrievalMetrics:  e.ComputeRetrievalMetrics(input),
		GenerationMetrics: e.ComputeGenerationMetrics(input),
	}
}

// DefaultEvaluator 默认评估器实例
var DefaultEvaluator = NewEvaluator()

// ComputeRetrievalMetrics 使用默认评估器计算检索指标
func ComputeRetrievalMetrics(input *MetricInput) RetrievalMetrics {
	return DefaultEvaluator.ComputeRetrievalMetrics(input)
}

// ComputeGenerationMetrics 使用默认评估器计算生成指标
func ComputeGenerationMetrics(input *MetricInput) GenerationMetrics {
	return DefaultEvaluator.ComputeGenerationMetrics(input)
}

// ComputeAll 使用默认评估器计算所有指标
func ComputeAll(input *MetricInput) MetricResult {
	return DefaultEvaluator.ComputeAll(input)
}
