package repository

type Banner struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

type BannersRepository interface {
	GetBanner(siteUrl string, slotId int) (Banner, error)
	AddSite(siteUrl string, slots []int) error
	AddBanner(bannerUrl string) error
	RemoveBanner(bannerUrl string) error
}
