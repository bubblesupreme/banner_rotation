package utils

import (
	"errors"
	"math"
	"math/rand"
)

var ErrNoStatistic = errors.New("insufficient data")

func Beta(x, y float64) float64 {
	return math.Gamma(x) * math.Gamma(y) / math.Gamma(x+y)
}

func SumFloat64(s []float64) float64 {
	res := 0.
	for _, v := range s {
		res += v
	}

	return res
}

func NormalizeFloat64(s []float64) ([]float64, error) {
	if len(s) == 0 {
		return nil, ErrNoStatistic
	}

	sum := SumFloat64(s)

	if sum == 0 {
		return nil, errors.New("failed to normalize, the sum is 0")
	}

	res := make([]float64, len(s), len(s))
	for i, v := range s {
		res[i] = v / sum
	}

	return res, nil
}

func DensityFloat64(s []float64) ([]float64, error) {
	if len(s) == 0 {
		return nil, ErrNoStatistic
	}

	norm, err := NormalizeFloat64(s)
	if err != nil {
		return nil, err
	}

	res := make([]float64, len(s), len(s))
	res[0] = norm[0]
	for i := 1; i < len(res); i++ {
		res[i] = res[i -1] + norm[i]
	}

	return res, nil
}

func ValIdxFromRatings(s []float64) (int, error) {
	if len(s) == 0 {
		return -1, ErrNoStatistic
	}

	density, err := DensityFloat64(s)
	if err != nil {
		return -1, err
	}

	p := rand.Float64()
	for i, d := range density {
		if p <= d {
			return i, nil
		}
	}

	return -1, errors.New("unexpected")
}
