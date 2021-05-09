package sqlrepository

import (
	"banner_rotation/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (r *sqlRepository) GetBanner(ctx context.Context, slotID int) (repository.Banner, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT banner_id FROM relations WHERE slot_id = $1;", slotID)
	if err != nil {
		return repository.Banner{}, err
	}
	defer func() {
		if err := rows.Err(); err != nil {
			log.Fatal("failed to check rows: ", err.Error())
		}
		if err := rows.Close(); err != nil {
			log.Fatal("failed to close rows: ", err.Error())
		}
	}()

	// TODO: multiarmed bandit
	bannerID := 0
	for rows.Next() {
		if err := rows.Scan(&bannerID); err != nil {
			return repository.Banner{}, err
		}
		break //nolint:staticcheck
	}

	return r.getBannerByID(ctx, bannerID)
}

func (r *sqlRepository) AddSlot(ctx context.Context) (repository.Slot, error) {
	slot := repository.Slot{}
	err := r.db.QueryRowContext(ctx, "INSERT INTO slots DEFAULT VALUES RETURNING id;").Scan(&slot.ID)
	if err != nil {
		return slot, err
	}

	log.WithFields(log.Fields{
		"id": slot.ID,
	}).Info("new slot was added")

	return slot, nil
}

func (r *sqlRepository) AddBanner(ctx context.Context, url string, description string) (repository.Banner, error) {
	banner := repository.Banner{}
	err := r.db.QueryRowContext(ctx, "SELECT id, url, description FROM banners WHERE url = $1 AND description = $2;", url, description).Scan(&banner.ID, &banner.URL, &banner.Description)
	if errors.Is(err, sql.ErrNoRows) {
		banner.URL = url
		banner.Description = description

		err := r.db.QueryRowContext(ctx, "INSERT INTO banners (url, description) VALUES ($1, $2) RETURNING id;", url, description).Scan(&banner.ID)
		if err != nil {
			return banner, err
		}

		log.WithFields(log.Fields{
			"url":         url,
			"description": description,
		}).Info("new banner was added")

		return banner, nil
	}
	if err != nil {
		return banner, err
	}

	log.WithFields(log.Fields{
		"url":             url,
		"description":     banner.Description,
		"new description": description,
	}).Warning("banner with the same url, description exists")

	return banner, nil
}

func (r *sqlRepository) RemoveBanner(ctx context.Context, bannerID int) error {
	result, resErr := r.db.ExecContext(ctx, "DELETE FROM banners WHERE id = $1", bannerID)
	if resErr == nil {
		logEntry := log.WithFields(log.Fields{
			"banner id": bannerID,
		})

		rows, err := result.RowsAffected()
		switch {
		case err != nil:
			log.Error("failed to check affected row while removing banner: ", err.Error())
		case rows == 0:
			logEntry.Warning("no banner to delete with the same url")
		case rows != 1:
			logEntry.Errorf("expected to affect 1 row, but affected %d while removing banner", rows)
		default:
			logEntry.Info("banner was removed")
		}
	}

	return resErr
}

func (r *sqlRepository) RemoveSlot(ctx context.Context, slotID int) error {
	result, resErr := r.db.ExecContext(ctx, "DELETE FROM slots WHERE id = $1", slotID)
	if resErr == nil {
		logEntry := log.WithFields(log.Fields{
			"slot id": slotID,
		})

		rows, err := result.RowsAffected()
		switch {
		case err != nil:
			log.Error("failed to check affected row while removing slot: ", err.Error())
		case rows == 0:
			logEntry.Warning("no slot to delete with the same id")
		case rows != 1:
			logEntry.Errorf("expected to affect 1 row, but affected %d while removing slot", rows)
		default:
			logEntry.Info("banner was removed")
		}
	}

	return resErr
}

func (r *sqlRepository) AddRelation(ctx context.Context, slotID int, bannerID int) error {
	if err := r.checkSlotAndBannerExistence(ctx, slotID, bannerID); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, "INSERT INTO relations (slot_id, banner_id, impressions, clicks) VALUES ($1, $2, 0, 0);", slotID, bannerID)
	if err == nil {
		log.WithFields(log.Fields{
			"slot id":   slotID,
			"banner id": bannerID,
		}).Info("new relation was added")
	}

	return nil
}

func (r *sqlRepository) RemoveRelation(ctx context.Context, slotID int, bannerID int) error {
	if err := r.checkSlotAndBannerExistence(ctx, slotID, bannerID); err != nil {
		return err
	}

	result, resErr := r.db.ExecContext(ctx, "DELETE FROM relations WHERE slot_id = $1 AND banner_id = $2", slotID, bannerID)
	if resErr == nil {
		logEntry := log.WithFields(log.Fields{
			"slot id":   slotID,
			"banner id": bannerID,
		})

		rows, err := result.RowsAffected()
		switch {
		case err != nil:
			log.Error("failed to check affected row while removing relation: ", err.Error())
		case rows == 0:
			logEntry.Warning("no relation to delete with the same slot id and banner url")
		case rows != 1:
			logEntry.Errorf("expected to affect 1 row, but affected %d while removing relation", rows)
		default:
			logEntry.Info("relation was removed")
		}
	}

	return resErr
}

func (r *sqlRepository) getBannerByID(ctx context.Context, bannerID int) (repository.Banner, error) {
	res := repository.Banner{}
	res.ID = bannerID
	err := r.db.QueryRowContext(ctx, "SELECT url, description FROM banners WHERE id = $1;", bannerID).Scan(&res.URL, &res.Description)
	if errors.Is(err, sql.ErrNoRows) {
		log.Errorf("no banner with id %d", bannerID)
	}

	return res, err
}

func (r *sqlRepository) checkSlotExistence(ctx context.Context, slotID int) (bool, error) {
	count := 0
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM slots WHERE id = $1;", slotID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err == nil && count > 1 {
		log.WithFields(log.Fields{
			"slot id": slotID,
		}).Warning("several slots with the same id")
	}

	return count > 0, err
}

func (r *sqlRepository) checkBannerExistenceByID(ctx context.Context, bannerID int) (bool, error) {
	count := 0
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM banners WHERE id = $1;", bannerID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err == nil && count > 1 {
		log.WithFields(log.Fields{
			"banner id": bannerID,
		}).Warning("several banners with the same id")
	}

	return count > 0, err
}

func (r *sqlRepository) Click(ctx context.Context, slotID int, bannerID int) error {
	if err := r.checkSlotAndBannerExistence(ctx, slotID, bannerID); err != nil {
		return err
	}

	result, resErr := r.db.ExecContext(ctx, "UPDATE relations SET clicks = clicks + 1 WHERE slot_id = $1 AND banner_id = $2", slotID, bannerID)
	if resErr == nil {
		rows, err := result.RowsAffected()

		if err != nil {
			log.Error("failed to check affected row while updating clicks: ", err.Error())
		} else if rows != 1 {
			log.WithFields(log.Fields{
				"slot id":   slotID,
				"banner id": bannerID,
			}).Errorf("expected to affect 1 row, but affected %d while updating clicks", rows)
		}
	}

	return resErr
}

func (r *sqlRepository) Show(ctx context.Context, slotID int, bannerID int) error {
	result, resErr := r.db.ExecContext(ctx, "UPDATE relations SET impressions = impressions + 1 WHERE slot_id = $1 AND banner_id = $2", slotID, bannerID)

	if resErr == nil {
		rows, err := result.RowsAffected()

		if err != nil {
			log.Error("failed to check affected row while updating impressions: ", err.Error())
		} else if rows != 1 {
			log.WithFields(log.Fields{
				"slot id":   slotID,
				"banner id": bannerID,
			}).Errorf("expected to affect 1 row, but affected %d while updating impressions", rows)
		}
	}

	return resErr
}

func (r *sqlRepository) GetAllBanners(ctx context.Context) ([]repository.Banner, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, url, description FROM banners;")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := rows.Err(); err != nil {
			log.Fatal("failed to check rows: ", err.Error())
		}
		if err := rows.Close(); err != nil {
			log.Fatal("failed to close rows: ", err.Error())
		}
	}()

	banners := make([]repository.Banner, 0)
	for rows.Next() {
		banner := repository.Banner{}
		if err := rows.Scan(&banner.ID, &banner.URL, &banner.Description); err != nil {
			log.Error("failed to scan row with banner while getting all banners")
		}
		banners = append(banners, banner)
	}

	return banners, nil
}

func (r *sqlRepository) checkSlotAndBannerExistence(ctx context.Context, slotID int, bannerID int) error {
	slotExist, err := r.checkSlotExistence(ctx, slotID)
	if err != nil {
		return err
	}
	if !slotExist {
		return fmt.Errorf("slot with id = %d doesn't exist", slotID)
	}

	bannerExist, err := r.checkBannerExistenceByID(ctx, bannerID)
	if err != nil {
		return err
	}
	if !bannerExist {
		return fmt.Errorf("banner with id = %d doesn't exist", bannerID)
	}

	return nil
}
