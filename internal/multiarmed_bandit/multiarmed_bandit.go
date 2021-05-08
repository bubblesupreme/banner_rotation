package multiarmedbandit

type BannerStatistic struct {
	BannerID int
	Impressions int
	Clicks int
}

type BannersStatistic = []BannerStatistic

type MultiarmedBandit interface {
	GetBanner(s BannersStatistic) (BannerStatistic, error)
}