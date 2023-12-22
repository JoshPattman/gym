package gym

import (
	"math"
	"math/rand"
)

func validateAction(action []float64, targetLength int) {
	if len(action) != targetLength {
		panic("Invalid action: length mismatch")
	}
	for _, v := range action {
		if v < -1 || v > 1 {
			panic("Invalid action: out of range -1 to 1")
		}
	}
}

// Gradient m=1 at x=0 and m=0 at x=inf, has max=1. Useful for mapping stuff like distances
func decay(x float64) float64 {
	if x < 0 {
		x = -x
		return -((-1 / (x + 1)) + 1)
	}
	return (-1 / (x + 1)) + 1
}

func clampAll(vals ...float64) []float64 {
	clampedVals := make([]float64, len(vals))
	for i := range vals {
		clampedVals[i] = vals[i]
		if clampedVals[i] < -1 {
			clampedVals[i] = -1
		}
		if clampedVals[i] > 1 {
			clampedVals[i] = 1
		}
	}
	return clampedVals
}

func normInt(std float64) int {
	rawRand := rand.NormFloat64() * std
	if rawRand < 0 {
		rawRand = -rawRand
	}
	return int(math.Round(rawRand))
}
