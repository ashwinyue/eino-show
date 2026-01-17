package metric

// PrecisionMetric calculates precision for retrieval evaluation
type PrecisionMetric struct{}

// NewPrecisionMetric creates a new PrecisionMetric instance
func NewPrecisionMetric() *PrecisionMetric {
	return &PrecisionMetric{}
}

// Compute calculates the precision score
func (r *PrecisionMetric) Compute(metricInput *MetricInput) float64 {
	gts := metricInput.RetrievalGT
	ids := metricInput.RetrievalIDs

	gtSets := SliceMap(gts, ToSet)
	ahit := Fold(gtSets, 0, func(a int, b map[int]struct{}) int { return a + Hit(ids, b) })

	if len(gts) == 0 {
		return 0.0
	}

	return float64(ahit) / float64(len(gts))
}
