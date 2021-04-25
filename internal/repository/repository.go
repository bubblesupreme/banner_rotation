package repository

type Banner struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type BannersRepository interface {
	GetBanner(siteURL string, slotID int) (Banner, error)
	AddSite(siteURL string, slots []int) error
	AddBanner(bannerURL string) error
	RemoveBanner(bannerURL string) error
}
