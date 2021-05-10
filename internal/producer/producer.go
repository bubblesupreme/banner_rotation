package producer

type Action struct {
	BannerID int `json:"banner"`
	SlotID   int `json:"slot"`
	GroupID  int `json:"group"`
}

type Producer interface {
	Show(a Action) error
	Click(a Action) error
	Shutdown() error
}
