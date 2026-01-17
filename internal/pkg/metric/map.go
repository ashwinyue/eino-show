package metric

// MAPMetric calculates Mean Average Precision for retrieval evaluation
type MAPMetric struct{}

// NewMAPMetric creates a new MAPMetric instance
func NewMAPMetric() *MAPMetric {
	return &MAPMetric{}
}

// Compute calculates the Mean Average Precision score
func (m *MAPMetric) Compute(metricInput *MetricInput) float64 {
	gts := metricInput.RetrievalGT
	ids := metricInput.RetrievalIDs

	gtSets := make([]map[int]struct{}, len(gts))
	for i, gt := range gts {
		gtSets[i] = make(map[int]struct{})
		for _, docID := range gt {
			gtSets[i][docID] = struct{}{}
		}
	}

	var apSum float64

	for _, gtSet := range gtSets {
		predHits := make([]bool, len(ids))
		for i, predID := range ids {
			if _, ok := gtSet[predID]; ok {
				predHits[i] = true
			} else {
				predHits[i] = false
			}
		}

		var (
			ap       float64
			hitCount int
		)

		for k := 0; k < len(predHits); k++ {
			if predHits[k] {
				hitCount++
				ap += float64(hitCount) / float64(k+1)
			}
		}
		if hitCount > 0 {
			ap /= float64(hitCount)
		}
		apSum += ap
	}

	if len(gtSets) == 0 {
		return 0
	}
	return apSum / float64(len(gtSets))
}
