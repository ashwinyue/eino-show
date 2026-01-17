package metric

// reference: https://github.com/dd-Rebecca/rouge

// RougeMetric implements ROUGE (Recall-Oriented Understudy for Gisting Evaluation) metrics
// for evaluating text summarization quality by comparing generated text to reference text
type RougeMetric struct {
	exclusive bool   // Whether to use exclusive matching mode
	metric    string // ROUGE metric type (e.g. "rouge-1", "rouge-l")
	stats     string // Statistic to return (e.g. "f", "p", "r")
}

// AvailableMetrics defines all supported ROUGE variants and their calculation functions
var AvailableMetrics = map[string]func([]string, []string, bool) map[string]float64{
	"rouge-1": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 1, false, exclusive)
	},
	"rouge-2": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 2, false, exclusive)
	},
	"rouge-3": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 3, false, exclusive)
	},
	"rouge-4": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 4, false, exclusive)
	},
	"rouge-5": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeN(hyp, ref, 5, false, exclusive)
	},
	"rouge-l": func(hyp, ref []string, exclusive bool) map[string]float64 {
		return rougeLSummaryLevel(hyp, ref, false, exclusive)
	},
}

// NewRougeMetric creates a new ROUGE metric calculator
func NewRougeMetric(exclusive bool, metrics, stats string) *RougeMetric {
	r := &RougeMetric{
		exclusive: exclusive,
		metric:    metrics,
		stats:     stats,
	}
	return r
}

// Compute calculates the ROUGE score between generated text and reference text
func (r *RougeMetric) Compute(metricInput *MetricInput) float64 {
	hyps := []string{metricInput.GeneratedTexts}
	refs := []string{metricInput.GeneratedGT}

	scores := 0.0
	count := 0

	for i := 0; i < len(hyps); i++ {
		hyp := splitSentences(hyps[i])
		ref := splitSentences(refs[i])

		fn := AvailableMetrics[r.metric]
		sc := fn(hyp, ref, r.exclusive)
		scores += sc[r.stats]

		count++
	}

	if count == 0 {
		return 0
	}
	return scores / float64(count)
}
