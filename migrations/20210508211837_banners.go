package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upBanners2, downBanners2)
}

func upBanners2(tx *sql.Tx) error {
	downBanners(tx)

	if _, err := tx.Exec(`
CREATE TABLE "banners" (
    "id" SERIAL NOT NULL,
    "url" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    PRIMARY KEY ("id")
);`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
CREATE TABLE "slots" (
    "id" SERIAL NOT NULL,
    PRIMARY KEY ("id")
);`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
CREATE TABLE "groups" (
    "id" SERIAL NOT NULL,
    "description" TEXT NOT NULL,
    PRIMARY KEY ("id")
);`); err != nil {
		return err
	}

	_, err := tx.Exec(`
CREATE TABLE "relations" (
    "id" SERIAL NOT NULL,
    "slot_id" INTEGER NOT NULL REFERENCES slots ON DELETE CASCADE,
    "banner_id" INTEGER NOT NULL REFERENCES banners ON DELETE CASCADE,
    "group_id" INTEGER NOT NULL REFERENCES groups ON DELETE CASCADE,
    "impressions" INTEGER NOT NULL,
    "clicks" INTEGER NOT NULL,
    PRIMARY KEY ("id")
);`)

	return err
}

func downBanners2(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE "banners";`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DROP TABLE "slots";`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DROP TABLE "relations";`); err != nil {
		return err
	}

	_, err := tx.Exec(`DROP TABLE "groups";`)

	return err
}
