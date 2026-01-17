package metric

import (
	"math"
)

// NDCGMetric calculates Normalized Discounted Cumulative Gain
type NDCGMetric struct {
	k int // Top k results to consider
}

// NewNDCGMetric creates a new NDCGMetric instance with given k value
func NewNDCGMetric(k int) *NDCGMetric {
	return &NDCGMetric{k: k}
}

// Compute calculates the NDCG score
func (n *NDCGMetric) Compute(metricInput *MetricInput) float64 {
	gts := metricInput.RetrievalGT
	ids := metricInput.RetrievalIDs

	if len(ids) > n.k {
		ids = ids[:n.k]
	}

	gtSets := make(map[int]struct{}, len(gts))
	countGt := 0
	for _, gt := range gts {
		countGt += len(gt)
		for _, g := range gt {
			gtSets[g] = struct{}{}
		}
	}

	relevanceScores := make(map[int]int)
	for _, docID := range ids {
		if _, exist := gtSets[docID]; exist {
			relevanceScores[docID] = 1
		} else {
			relevanceScores[docID] = 0
		}
	}

	var dcg float64
	for i, docID := range ids {
		dcg += (math.Pow(2, float64(relevanceScores[docID])) - 1) / math.Log2(float64(i+2))
	}

	idealLen := min(countGt, len(ids))
	idealPred := make([]int, len(ids))
	for i := 0; i < len(ids); i++ {
		if i < idealLen {
			idealPred[i] = 1
		} else {
			idealPred[i] = 0
		}
	}

	var idcg float64
	for i, relevance := range idealPred {
		idcg += float64(relevance) / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}
