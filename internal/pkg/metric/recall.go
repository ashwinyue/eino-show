package metric

// RecallMetric calculates recall for retrieval evaluation
type RecallMetric struct{}

// NewRecallMetric creates a new RecallMetric instance
func NewRecallMetric() *RecallMetric {
	return &RecallMetric{}
}

// Compute calculates the recall score
func (r *RecallMetric) Compute(metricInput *MetricInput) float64 {
	gts := metricInput.RetrievalGT
	ids := metricInput.RetrievalIDs

	gtSets := SliceMap(gts, ToSet)
	ahit := Fold(gtSets, 0, func(a int, b map[int]struct{}) int { return a + Hit(ids, b) })

	if len(gtSets) == 0 {
		return 0.0
	}

	return float64(ahit) / float64(len(gtSets))
}
