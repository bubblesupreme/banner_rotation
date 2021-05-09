package multiarmedbandit

import (
	"banner_rotation/utils"
)

type thompsonBandit struct {
	minEvents int
}

func NewThompsonBandit(minEvents int) (MultiarmedBandit, error) {
	return &thompsonBandit{
		minEvents: minEvents,
	}, nil
}

func (t *thompsonBandit) GetBanner(s BannersStatistic) (BannerStatistic, error) {
	if s == nil {
		return BannerStatistic{}, utils.ErrNoStatistic
	}

	warm := warmBanners(s, t.minEvents)
	ratings := calculateRatings(s, warm)
	maxIdx, err := chooseRating(ratings)
	if err != nil {
		return BannerStatistic{}, err
	}

	return s[maxIdx], nil
}

func warmBanners(s BannersStatistic, minActions int) []int {
	warm := make([]int, 0)
	for i, b := range s {
		if b.Impressions >= minActions {
			warm = append(warm, i)
		}
	}

	return warm
}

func calculateRatings(s BannersStatistic, warm []int) []float64 {
	ratings := make([]float64, len(s))
	warmIdx := -1
	if len(warm) > 0 {
		warmIdx = 0
	}
	for i, b := range s {
		if warmIdx != -1 && warm[warmIdx] == i {
			warmIdx++
			ratings[i] = float64(b.Clicks) / float64(b.Impressions)
		} else { // banner is cold
			ratings[i] = 1.
		}
	}

	return ratings
}

func chooseRating(ratings []float64) (int, error) {
	return utils.ValIdxFromRatings(ratings)
}
