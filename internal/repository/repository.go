package repository

import "context"

type Banner struct {
	ID          int    `json:"id"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Slot struct {
	ID int `json:"slot"`
}

type BannersRepository interface {
	GetBanner(ctx context.Context, slotID int) (Banner, error)
	AddSlot(ctx context.Context) (Slot, error)
	AddBanner(ctx context.Context, bannerURL string, description string) (Banner, error)
	AddRelation(ctx context.Context, slotID int, bannerID int) error
	RemoveBanner(ctx context.Context, bannerID int) error
	RemoveSlot(ctx context.Context, slotID int) error
	RemoveRelation(ctx context.Context, slotID int, bannerID int) error
	Click(ctx context.Context, slotID int, bannerID int) error
	GetAllBanners(ctx context.Context) ([]Banner, error)
}
