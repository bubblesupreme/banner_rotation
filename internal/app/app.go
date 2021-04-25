package app

import (
	"encoding/json"
	"net/http"

	"banner_rotation/internal/repository"

	log "github.com/sirupsen/logrus"
)

type BannersApp struct {
	repo repository.BannersRepository
}

func NewBannersApp(repo repository.BannersRepository) *BannersApp {
	return &BannersApp{
		repo: repo,
	}
}

func (a *BannersApp) GetBanner(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		SiteUrl string `json:"site"`
		SlotId  int    `json:"slot"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error("failed to parse request parameters: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	banner, err := a.repo.GetBanner(reqData.SiteUrl, reqData.SlotId)
	if err != nil {
		log.WithFields(log.Fields{
			"site":    reqData.SiteUrl,
			"slot id": reqData.SlotId,
		}).Error("failed to get banner: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.NewEncoder(w).Encode(&banner); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *BannersApp) AddSite(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		SiteUrl string `json:"site"`
		SlotIds []int  `json:"slot ids"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error("failed to parse request parameters: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := a.repo.AddSite(reqData.SiteUrl, reqData.SlotIds)
	if err != nil {
		log.WithFields(log.Fields{
			"site":     reqData.SiteUrl,
			"slot ids": reqData.SlotIds,
		}).Error("failed to add new site: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *BannersApp) AddBanner(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		BannerUrl string `json:"banner"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Error("failed to parse request parameters: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := a.repo.AddBanner(reqData.BannerUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"banner url": reqData.BannerUrl,
		}).Error("failed to add new banner: ", err.Error())

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
