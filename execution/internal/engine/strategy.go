package engine

import "trading-bot/execution/internal/types"

type Strategy interface {
	Name() string
	Evaluate(state *types.MarketState) *types.Decision
}

type PriceHistory struct {
	prices []float64
	maxLen int
	pos    int
	count  int
}

func NewPriceHistory(maxLen int) *PriceHistory {
	return &PriceHistory{
		prices: make([]float64, maxLen),
		maxLen: maxLen,
	}
}

func (h *PriceHistory) Add(price float64) {
	if price <= 0 {
		return
	}
	h.prices[h.pos] = price
	h.pos = (h.pos + 1) % h.maxLen
	if h.count < h.maxLen {
		h.count++
	}
}

func (h *PriceHistory) Values() []float64 {
	if h.count == 0 {
		return nil
	}
	result := make([]float64, h.count)
	for i := 0; i < h.count; i++ {
		idx := (h.pos - h.count + i + h.maxLen) % h.maxLen
		result[i] = h.prices[idx]
	}
	return result
}

func (h *PriceHistory) Len() int {
	return h.count
}

func (h *PriceHistory) Mean() float64 {
	vals := h.Values()
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func (h *PriceHistory) StdDev() float64 {
	vals := h.Values()
	if len(vals) < 2 {
		return 0
	}
	mean := h.Mean()
	sum := 0.0
	for _, v := range vals {
		diff := v - mean
		sum += diff * diff
	}
	variance := sum / float64(len(vals))
	return sqrt(variance)
}

func (h *PriceHistory) RSI(period int) float64 {
	vals := h.Values()
	if len(vals) < period+1 {
		return 50
	}
	avgGain, avgLoss := 0.0, 0.0
	start := len(vals) - period - 1
	for i := start + 1; i < len(vals); i++ {
		change := vals[i] - vals[i-1]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

func (h *PriceHistory) BollingerBands(period int, stdDev float64) (middle, upper, lower float64) {
	if h.count < period {
		return 0, 0, 0
	}
	window := h.Values()[h.count-period:]
	mean := 0.0
	for _, v := range window {
		mean += v
	}
	middle = mean / float64(len(window))

	sumSq := 0.0
	for _, v := range window {
		diff := v - middle
		sumSq += diff * diff
	}
	band := stdDev * sqrt(sumSq/float64(len(window)))
	upper = middle + band
	lower = middle - band
	return
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
