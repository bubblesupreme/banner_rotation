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

type Group struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type BannersRepository interface {
	GetBanner(ctx context.Context, slotID, groupID int) (Banner, error)
	AddSlot(ctx context.Context) (Slot, error)
	AddBanner(ctx context.Context, bannerURL, description string) (Banner, error)
	AddRelation(ctx context.Context, slotID, bannerID int) error
	RemoveBanner(ctx context.Context, bannerID int) error
	RemoveSlot(ctx context.Context, slotID int) error
	RemoveRelation(ctx context.Context, slotID, bannerID int) error
	Click(ctx context.Context, slotID, bannerID, groupID int) error
	GetAllBanners(ctx context.Context) ([]Banner, error)
	AddGroup(ctx context.Context, description string) (Group, error)
	RemoveGroup(ctx context.Context, groupID int) error
	GetAllGroups(ctx context.Context) ([]Group, error)
	Show(ctx context.Context, slotID int, bannerID int, groupID int) error
}
