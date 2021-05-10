package app

import (
	"banner_rotation/internal/producer"
	"banner_rotation/internal/repository"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type BannersApp struct {
	repo     repository.BannersRepository
	producer producer.Producer
}

func NewBannersApp(repo repository.BannersRepository, producer producer.Producer) *BannersApp {
	return &BannersApp{
		repo:     repo,
		producer: producer,
	}
}

func (a *BannersApp) GetBanner(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	reqData := struct {
		SlotID  int `json:"slot"`
		GroupID int `json:"group"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	banner, err := a.repo.GetBanner(r.Context(), reqData.SlotID, reqData.GroupID)
	if err != nil {
		log.WithFields(log.Fields{
			"slot id":  reqData.SlotID,
			"group id": reqData.GroupID,
		}).Error("failed to get banner: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(&banner); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) AddSlot(w http.ResponseWriter, r *http.Request) {
	slot, err := a.repo.AddSlot(r.Context())
	if err != nil {
		log.Error("failed to add new slot: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(&slot); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) AddBanner(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	reqData := struct {
		BannerURL   string `json:"url"`
		BannerDescr string `json:"description"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	banner, err := a.repo.AddBanner(r.Context(), reqData.BannerURL, reqData.BannerDescr)
	if err != nil {
		log.WithFields(log.Fields{
			"url":         reqData.BannerURL,
			"description": reqData.BannerDescr,
		}).Error("failed to add new banner: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(&banner); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) AddRelation(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		SlotID   int `json:"slot"`
		BannerID int `json:"banner"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.repo.AddRelation(r.Context(), reqData.SlotID, reqData.BannerID); err != nil {
		log.WithFields(log.Fields{
			"slot id":   reqData.SlotID,
			"banner id": reqData.BannerID,
		}).Error("failed to add new relation: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) RemoveBanner(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		BannerID int `json:"banner"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.repo.RemoveBanner(r.Context(), reqData.BannerID); err != nil {
		log.WithFields(log.Fields{
			"banner id": reqData.BannerID,
		}).Error("failed to remove banner: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) RemoveSlot(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		SlotID int `json:"slot"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.repo.RemoveSlot(r.Context(), reqData.SlotID); err != nil {
		log.WithFields(log.Fields{
			"slot id": reqData.SlotID,
		}).Error("failed to remove banner: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) RemoveRelation(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		SlotID   int `json:"slot"`
		BannerID int `json:"banner"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.repo.RemoveRelation(r.Context(), reqData.SlotID, reqData.BannerID); err != nil {
		log.WithFields(log.Fields{
			"slot id":   reqData.SlotID,
			"banner id": reqData.BannerID,
		}).Error("failed to remove relation: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) Click(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	reqData := struct {
		SlotID   int `json:"slot"`
		BannerID int `json:"banner"`
		GroupID  int `json:"group"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logEntry := log.WithFields(log.Fields{
		"slot id":   reqData.SlotID,
		"banner id": reqData.BannerID,
		"group id":  reqData.GroupID,
	})
	if err := a.repo.Click(r.Context(), reqData.SlotID, reqData.BannerID, reqData.GroupID); err != nil {
		logEntry.Error("failed to count the click: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := a.producer.Click(producer.Action{
		BannerID: reqData.BannerID,
		SlotID:   reqData.SlotID,
		GroupID:  reqData.GroupID,
	}); err != nil {
		logEntry.Error("failed to publish the click action: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) Show(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	reqData := struct {
		SlotID   int `json:"slot"`
		BannerID int `json:"banner"`
		GroupID  int `json:"group"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logEntry := log.WithFields(log.Fields{
		"slot id":   reqData.SlotID,
		"banner id": reqData.BannerID,
		"group id":  reqData.GroupID,
	})
	if err := a.repo.Show(r.Context(), reqData.SlotID, reqData.BannerID, reqData.GroupID); err != nil {
		logEntry.Error("failed to count the showing: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := a.producer.Show(producer.Action{
		BannerID: reqData.BannerID,
		SlotID:   reqData.SlotID,
		GroupID:  reqData.GroupID,
	}); err != nil {
		logEntry.Error("failed to publish the show action: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) GetAllBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := a.repo.GetAllBanners(r.Context())
	if err != nil {
		log.Error("failed to get all available banners: ", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(&banners); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *BannersApp) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := a.repo.GetAllGroups(r.Context())
	if err != nil {
		log.Error("failed to get all available social groups: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(&groups); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *BannersApp) AddGroup(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		GroupDescr string `json:"description"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	group, err := a.repo.AddGroup(r.Context(), reqData.GroupDescr)
	if err != nil {
		log.WithFields(log.Fields{
			"description": reqData.GroupDescr,
		}).Error("failed to add new social group: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(&group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *BannersApp) RemoveGroup(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		GroupID int `json:"group"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error(parseRequestParamsErr(err))

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.repo.RemoveGroup(r.Context(), reqData.GroupID); err != nil {
		log.WithFields(log.Fields{
			"group id": reqData.GroupID,
		}).Error("failed to remove group: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func parseRequestParamsErr(err error) string {
	return "failed to parse request parameters: " + err.Error()
}
