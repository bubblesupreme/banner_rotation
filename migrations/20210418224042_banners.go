package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upBanners, downBanners)
}

func upBanners(tx *sql.Tx) error {
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

	_, err := tx.Exec(`
CREATE TABLE "relations" (
    "id" SERIAL NOT NULL,
    "slot_id" INTEGER NOT NULL,
    "banner_id" INTEGER NOT NULL,
    "impressions" INTEGER NOT NULL,
    "clicks" INTEGER NOT NULL,
    PRIMARY KEY ("id")
);`)

	return err
}

func downBanners(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE "banners";`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DROP TABLE "slots";`); err != nil {
		return err
	}

	_, err := tx.Exec(`DROP TABLE "relations";`)

	return err
}
