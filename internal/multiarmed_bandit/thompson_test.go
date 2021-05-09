package multiarmedbandit

import (
	"banner_rotation/utils"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThompsonAllCold(t *testing.T) {
	minEvents := 50
	nRun := 1000
	nBanners := 50

	bandit, err := NewThompsonBandit(minEvents)
	assert.NoError(t, err)

	s := make([]BannerStatistic, nBanners)
	for i := 0; i < len(s); i++ {
		s[i].BannerID = i
		s[i].Impressions = rand.Intn(minEvents)
		if s[i].Impressions > 0 {
			s[i].Clicks = rand.Intn(s[i].Impressions)
		} else {
			s[i].Clicks = 0
		}
	}

	choices := make([]int, len(s))
	for i := 0; i < nRun; i++ {
		b, err := bandit.GetBanner(s)
		assert.NoError(t, err)
		choices[b.BannerID]++
	}

	for _, ch := range choices {
		assert.True(t, ch > 0)
	}
}

func TestThompsonAllWarm(t *testing.T) {
	minEvents := 50
	nRun := 50000
	nBanners := 50
	checkIdx := 13

	bandit, err := NewThompsonBandit(minEvents)
	assert.NoError(t, err)

	s := make([]BannerStatistic, nBanners)
	for i := 0; i < len(s); i++ {
		s[i].BannerID = i
		s[i].Impressions = rand.Intn(minEvents) + minEvents
		s[i].Clicks = s[i].Impressions / 2
	}
	s[checkIdx].Impressions = minEvents * 2
	s[checkIdx].Clicks = s[checkIdx].Impressions

	choices := make([]int, len(s))
	for i := 0; i < nRun; i++ {
		b, err := bandit.GetBanner(s)
		assert.NoError(t, err)
		choices[b.BannerID]++
	}

	maxIdx := 0
	for i, ch := range choices {
		if ch > choices[maxIdx] {
			maxIdx = i
		}
	}
	assert.True(t, maxIdx == checkIdx)
}

func TestThompsonOneCold(t *testing.T) {
	minEvents := 50
	nRun := 50000
	nBanners := 50
	checkIdx := 13

	bandit, err := NewThompsonBandit(minEvents)
	assert.NoError(t, err)

	s := make([]BannerStatistic, nBanners)
	for i := 0; i < len(s); i++ {
		s[i].BannerID = i
		s[i].Impressions = rand.Intn(minEvents) + minEvents
		s[i].Clicks = s[i].Impressions / 2
	}
	s[checkIdx].Impressions = rand.Intn(minEvents)
	s[checkIdx].Clicks = rand.Intn(minEvents)

	choices := make([]int, len(s))
	for i := 0; i < nRun; i++ {
		b, err := bandit.GetBanner(s)
		assert.NoError(t, err)
		choices[b.BannerID]++
	}

	assert.True(t, choices[checkIdx] > 0)
}

func TestThompsonOneValue(t *testing.T) {
	minEvents := 50
	nBanners := 1

	bandit, err := NewThompsonBandit(minEvents)
	assert.NoError(t, err)

	s := make([]BannerStatistic, nBanners)
	for i := 0; i < len(s); i++ {
		s[i].BannerID = i
		s[i].Impressions = rand.Intn(minEvents) + minEvents
		s[i].Clicks = s[i].Impressions / 2
	}

	b, err := bandit.GetBanner(s)
	assert.NoError(t, err)
	assert.True(t, b.BannerID == 0)
}

func TestThompsonNoStatistic(t *testing.T) {
	minEvents := 50

	bandit, err := NewThompsonBandit(minEvents)
	assert.NoError(t, err)

	_, err = bandit.GetBanner(nil)
	assert.EqualError(t, err, utils.ErrNoStatistic.Error())
}
