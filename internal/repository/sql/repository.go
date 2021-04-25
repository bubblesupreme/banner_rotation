package sqlrepository

import (
	"database/sql"

	"banner_rotation/internal/repository"
	log "github.com/sirupsen/logrus"
)

type sqlRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) repository.BannersRepository {
	return &sqlRepository{
		db: db,
	}
}

func (r *sqlRepository) GetBanner(site string, slotID int) (repository.Banner, error) {
	siteID, err := r.getSiteID(site)
	if err != nil {
		log.WithFields(log.Fields{
			"site url": site,
		}).Error("failed to find site id by url: ", err.Error())

		return repository.Banner{}, err
	}

	return r.getBanner(siteID, slotID)
}

func (r *sqlRepository) AddSite(siteURL string, slots []int) error {
	return nil
}

func (r *sqlRepository) AddBanner(bannerURL string) error {
	return nil
}

func (r *sqlRepository) RemoveBanner(bannerURL string) error {
	return nil
}

func (r *sqlRepository) getBanner(siteID int, slotID int) (repository.Banner, error) {
	res := repository.Banner{}

	rows, err := r.db.Query("SELECT banner_id FROM banners WHERE site_id = $1 AND slot_id = $2;", siteID, slotID)
	if err != nil {
		return res, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Fatal("failed to close rows: ", err.Error())
		}
		if err := rows.Err(); err != nil {
			log.Fatal("failed to check rows: ", err.Error())
		}
	}()

	// TODO: multiarmed bandit
	for rows.Next() {
		if err := rows.Scan(&res.ID); err != nil {
			return res, err
		}
		break //nolint:staticcheck
	}

	res.URL, err = r.getBannerURL(res.ID)
	return res, err
}

// func (r *sqlRepository) getSiteURL(siteID int) (string, error) {
//	row := r.db.QueryRow("SELECT url FROM site_urls WHERE id = $1;", siteID)
//	if row.Err() != nil {
//		return "", row.Err()
//	}
//
//	res := ""
//	if err := row.Scan(&res); err != nil {
//		return "", err
//	}
//	return res, nil
// }

func (r *sqlRepository) getSiteID(siteURL string) (int, error) {
	row := r.db.QueryRow("SELECT id FROM site_urls WHERE url = $1;", siteURL)
	if row.Err() != nil {
		return -1, row.Err()
	}

	res := -1
	if err := row.Scan(&res); err != nil {
		return -1, err
	}
	return res, nil
}

func (r *sqlRepository) getBannerURL(bannerID int) (string, error) {
	row := r.db.QueryRow("SELECT url FROM banner_urls WHERE id = $1;", bannerID)
	if row.Err() != nil {
		return "", row.Err()
	}

	res := ""
	if err := row.Scan(&res); err != nil {
		return "", err
	}
	return res, nil
}
