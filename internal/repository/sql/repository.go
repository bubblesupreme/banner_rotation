package sqlrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	bandit "github.com/bubblesupreme/banner_rotation/internal/multiarmed_bandit"
	"github.com/bubblesupreme/banner_rotation/internal/repository"

	log "github.com/sirupsen/logrus"
)

type sqlRepository struct {
	db     *sql.DB
	bandit bandit.MultiarmedBandit
}

type relation struct {
	slotID   int
	bannerID int
}

func NewSQLRepository(db *sql.DB, bandit bandit.MultiarmedBandit) repository.BannersRepository {
	return &sqlRepository{
		db:     db,
		bandit: bandit,
	}
}

func (r *sqlRepository) GetBanner(ctx context.Context, slotID, groupID int) (repository.Banner, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT banner_id, impressions, clicks FROM relations WHERE slot_id = $1 AND group_id = $2;", slotID, groupID) //nolint:rowserrcheck,sqlclosecheck
	if err != nil {
		return repository.Banner{}, err
	}
	defer checkRows(rows)

	s := bandit.BannersStatistic{}
	b := bandit.BannerStatistic{}
	for rows.Next() {
		if err := rows.Scan(&b.BannerID, &b.Impressions, &b.Clicks); err != nil {
			return repository.Banner{}, err
		}
		s = append(s, b)
	}

	if len(s) == 0 {
		log.WithFields(log.Fields{
			"slot id":  slotID,
			"group id": groupID,
		}).Error("no one rows were returned from sql")
		return repository.Banner{}, fmt.Errorf("no one banner relations for given parameters")
	}

	banner, err := r.bandit.GetBanner(s)
	if err != nil {
		return repository.Banner{}, err
	}

	res, err := r.getBannerByID(ctx, banner.BannerID)
	if err != nil {
		return res, err
	}

	log.WithFields(log.Fields{
		"slot id":            slotID,
		"group id":           groupID,
		"banner id":          res.ID,
		"banner url":         res.URL,
		"banner description": res.Description,
	}).Info("get banner function")
	return res, nil
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
		"id":          banner.ID,
		"url":         banner.URL,
		"description": banner.Description,
	}).Warning("banner exists")

	return banner, nil
}

func (r *sqlRepository) RemoveBanner(ctx context.Context, bannerID int) error {
	result, resErr := r.db.ExecContext(ctx, "DELETE FROM banners WHERE id = $1;", bannerID)
	if resErr == nil {
		logEntry := log.WithFields(log.Fields{
			"banner id": bannerID,
		})

		rows, err := result.RowsAffected()
		switch {
		case err != nil:
			log.Error("failed to check affected row while removing banner: ", err.Error())
		case rows == 0:
			logEntry.Warning("no banner to delete with the same id")
		case rows != 1:
			logEntry.Errorf("expected to affect 1 row, but affected %d while removing banner", rows)
		default:
			logEntry.Info("banner was removed")
		}
	}

	return resErr
}

func (r *sqlRepository) RemoveSlot(ctx context.Context, slotID int) error {
	result, resErr := r.db.ExecContext(ctx, "DELETE FROM slots WHERE id = $1;", slotID)
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
			logEntry.Info("slot was removed")
		}
	}

	return resErr
}

func (r *sqlRepository) AddRelation(ctx context.Context, slotID int, bannerID int) error {
	if err := r.checkSlotBannerExistence(ctx, slotID, bannerID); err != nil {
		return err
	}

	relationExist, err := r.checkRelationExistence(ctx, slotID, bannerID)
	if err != nil {
		return err
	}

	logEntry := log.WithFields(log.Fields{
		"slot id":   slotID,
		"banner id": bannerID,
	})
	if relationExist {
		logEntry.Warning("relation exists")
		return nil
	}

	groups, err := r.GetAllGroups(ctx)
	if err != nil {
		logEntry.Error("failed to get all groups while adding new relation")

		return err
	}

	if len(groups) == 0 {
		return fmt.Errorf("no one group exists")
	}

	for _, g := range groups {
		_, err := r.db.ExecContext(ctx, "INSERT INTO relations (slot_id, banner_id, group_id, impressions, clicks) VALUES ($1, $2, $3, 0, 0);", slotID, bannerID, g.ID)
		if err == nil {
			log.WithFields(log.Fields{
				"slot id":   slotID,
				"banner id": bannerID,
				"group id":  g.ID,
			}).Info("new relation was added")
		}
	}

	return nil
}

func (r *sqlRepository) RemoveRelation(ctx context.Context, slotID int, bannerID int) error {
	result, resErr := r.db.ExecContext(ctx, "DELETE FROM relations WHERE slot_id = $1 AND banner_id = $2;", slotID, bannerID)
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

func (r *sqlRepository) checkGroupExistence(ctx context.Context, groupID int) (bool, error) {
	count := 0
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM groups WHERE id = $1;", groupID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err == nil && count > 1 {
		log.WithFields(log.Fields{
			"group id": groupID,
		}).Warning("several groups with the same id")
	}

	return count > 0, err
}

func (r *sqlRepository) Click(ctx context.Context, slotID, bannerID, groupID int) error {
	if err := r.checkFullRelationExistence(ctx, slotID, bannerID, groupID); err != nil {
		return err
	}

	result, resErr := r.db.ExecContext(ctx, "UPDATE relations SET clicks = clicks + 1 WHERE slot_id = $1 AND banner_id = $2 AND group_id = $3;", slotID, bannerID, groupID)
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

func (r *sqlRepository) Show(ctx context.Context, slotID, bannerID, groupID int) error {
	if err := r.checkFullRelationExistence(ctx, slotID, bannerID, groupID); err != nil {
		return err
	}

	result, resErr := r.db.ExecContext(ctx, "UPDATE relations SET impressions = impressions + 1 WHERE slot_id = $1 AND banner_id = $2 AND group_id = $3;", slotID, bannerID, groupID)

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
	rows, err := r.db.QueryContext(ctx, "SELECT id, url, description FROM banners;") //nolint:rowserrcheck,sqlclosecheck
	if err != nil {
		log.Fatal(err)
	}
	defer checkRows(rows)

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

func (r *sqlRepository) checkSlotBannerExistence(ctx context.Context, slotID, bannerID int) error {
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

func (r *sqlRepository) checkRelationExistence(ctx context.Context, slotID, bannerID int) (bool, error) {
	count := 0
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM relations WHERE slot_id = $1 AND banner_id = $2;", slotID, bannerID).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return count > 0, err
}

func (r *sqlRepository) checkFullRelationExistence(ctx context.Context, slotID, bannerID, groupID int) error {
	if err := r.checkSlotBannerGroupExistence(ctx, slotID, bannerID, groupID); err != nil {
		return err
	}

	relationExist, err := r.checkRelationExistence(ctx, slotID, bannerID)
	if err != nil {
		return err
	}

	if !relationExist {
		return fmt.Errorf("relation with slot id = %d and banner id = %d doesn't exist", slotID, bannerID)
	}

	return nil
}

func (r *sqlRepository) checkSlotBannerGroupExistence(ctx context.Context, slotID, bannerID, groupID int) error {
	if err := r.checkSlotBannerExistence(ctx, slotID, bannerID); err != nil {
		return err
	}

	groupExist, err := r.checkGroupExistence(ctx, groupID)
	if err != nil {
		return err
	}
	if !groupExist {
		return fmt.Errorf("group with id = %d doesn't exist", groupID)
	}

	return nil
}

func (r *sqlRepository) AddGroup(ctx context.Context, description string) (repository.Group, error) {
	group := repository.Group{}
	err := r.db.QueryRowContext(ctx, "SELECT id, description FROM banners WHERE description = $1;", description).Scan(&group.ID, &group.Description)
	if errors.Is(err, sql.ErrNoRows) {
		group.Description = description

		err := r.db.QueryRowContext(ctx, "INSERT INTO groups (description) VALUES ($1) RETURNING id;", description).Scan(&group.ID)
		if err != nil {
			return group, err
		}

		log.WithFields(log.Fields{
			"description": description,
		}).Info("new social group was added")

		return group, r.addGroupToRelation(ctx, group.ID)
	}
	if err != nil {
		return group, err
	}

	log.WithFields(log.Fields{
		"id":          group.ID,
		"description": group.Description,
	}).Warning("social group exists")

	return group, nil
}

func (r *sqlRepository) addGroupToRelation(ctx context.Context, groupID int) error {
	rows, err := r.db.QueryContext(ctx, "SELECT DISTINCT slot_id, banner_id FROM relations;") //nolint:rowserrcheck,sqlclosecheck
	if err != nil {
		return err
	}
	defer checkRows(rows)

	rel := relation{}
	relations := make([]relation, 0)
	for rows.Next() {
		if err := rows.Scan(&rel.slotID, &rel.bannerID); err != nil {
			return err
		}
		relations = append(relations, rel)
	}

	for _, relation := range relations {
		_, err := r.db.ExecContext(ctx, "INSERT INTO relations (slot_id, banner_id, group_id, impressions, clicks) VALUES ($1, $2, $3, 0, 0);", relation.slotID, relation.bannerID, groupID)
		logEntry := log.WithFields(log.Fields{
			"slot id":   relation.slotID,
			"banner id": relation.bannerID,
			"group id":  groupID,
		})
		if err == nil {
			logEntry.Info("new relation was added")
		} else {
			logEntry.Error("failed to add new relation")
		}
	}
	return nil
}

func (r *sqlRepository) RemoveGroup(ctx context.Context, groupID int) error {
	result, resErr := r.db.ExecContext(ctx, "DELETE FROM groups WHERE id = $1;", groupID)
	if resErr == nil {
		logEntry := log.WithFields(log.Fields{
			"group id": groupID,
		})

		rows, err := result.RowsAffected()
		switch {
		case err != nil:
			log.Error("failed to check affected row while removing group: ", err.Error())
		case rows == 0:
			logEntry.Warning("no group to delete with the same id")
		case rows != 1:
			logEntry.Errorf("expected to affect 1 row, but affected %d while removing group", rows)
		default:
			logEntry.Info("group was removed")
		}
	}

	return resErr
}

func (r *sqlRepository) GetAllGroups(ctx context.Context) ([]repository.Group, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, description FROM groups;") //nolint:rowserrcheck,sqlclosecheck
	if err != nil {
		log.Fatal(err)
	}
	defer checkRows(rows)

	groups := make([]repository.Group, 0)
	for rows.Next() {
		group := repository.Group{}
		if err := rows.Scan(&group.ID, &group.Description); err != nil {
			log.Error("failed to scan row with group while getting all groups")
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func checkRows(rows *sql.Rows) {
	if err := rows.Err(); err != nil {
		log.Fatal("failed to check rows: ", err.Error())
	}
	if err := rows.Close(); err != nil {
		log.Fatal("failed to close rows: ", err.Error())
	}
}
