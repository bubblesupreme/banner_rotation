package sql_repository

import (
	"banner_rotation/internal/repository"
	"database/sql"

	log "github.com/sirupsen/logrus"
)

type sqlRepository struct {
	db *sql.DB
}

func NewSqlRepository(db *sql.DB) repository.BannersRepository {
	return &sqlRepository{
		db: db,
	}
}

func (r *sqlRepository) GetBanner(site string, slotId int) (repository.Banner, error) {
	siteId, err := r.getSiteId(site)
	if err != nil {
		log.WithFields(log.Fields{
			"site url": site,
		}).Error("failed to find site id by url: ", err.Error())

		return repository.Banner{}, err
	}

	return r.getBanner(siteId, slotId)
}

func (r *sqlRepository) AddSite(siteUrl string, slots []int) error {
	return nil
}

func (r *sqlRepository) AddBanner(bannerUrl string) error {
	return nil
}

func (r *sqlRepository) RemoveBanner(bannerUrl string) error {
	return nil
}

func (r *sqlRepository) getBanner(siteId int, slotId int) (repository.Banner, error) {
	res := repository.Banner{}

	rows, err := r.db.Query("SELECT banner_id FROM banners WHERE site_id = $1 AND slot_id = $2;", siteId, slotId)
	if err != nil {
		return res, err
	}

	for rows.Next() {
		if err := rows.Scan(&res.Id); err != nil {
			return res, err
		}
		break // return the first
	}

	res.Url, err = r.getBannerUrl(res.Id)
	return res, err
}

func (r *sqlRepository) getSiteUrl(siteId int) (string, error) {
	row := r.db.QueryRow("SELECT url FROM site_urls WHERE id = $1;", siteId)
	if row.Err() != nil {
		return "", row.Err()
	}

	res := ""
	if err := row.Scan(&res); err != nil {
		return "", err
	}
	return res, nil
}

func (r *sqlRepository) getSiteId(siteUrl string) (int, error) {
	row := r.db.QueryRow("SELECT id FROM site_urls WHERE url = $1;", siteUrl)
	if row.Err() != nil {
		return -1, row.Err()
	}

	res := -1
	if err := row.Scan(&res); err != nil {
		return -1, err
	}
	return res, nil
}

func (r *sqlRepository) getBannerUrl(bannerId int) (string, error) {
	row := r.db.QueryRow("SELECT url FROM banner_urls WHERE id = $1;", bannerId)
	if row.Err() != nil {
		return "", row.Err()
	}

	res := ""
	if err := row.Scan(&res); err != nil {
		return "", err
	}
	return res, nil
}
