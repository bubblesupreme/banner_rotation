package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upBanners, downBanners)
}

func upBanners(tx *sql.Tx) error {
	if _, err := tx.Exec(`CREATE TABLE "banner_urls" (
    "id" SERIAL NOT NULL,
    "url" TEXT NOT NULL,
    PRIMARY KEY ("id")
);`); err != nil {
		return err
	}

	if _, err := tx.Exec(`CREATE TABLE "site_urls" (
    "id" SERIAL NOT NULL,
    "url" TEXT NOT NULL,
    PRIMARY KEY ("id")
);`); err != nil {
		return err
	}

	_, err := tx.Exec(`CREATE TABLE "banners" (
    "id" SERIAL NOT NULL,
    "banner_id" INTEGER NOT NULL,
    "site_id" INTEGER NOT NULL,
    "slot_id" INTEGER NOT NULL,
    PRIMARY KEY ("id")
);`)

	return err
}

func downBanners(tx *sql.Tx) error {
	if _, err := tx.Exec(`DROP TABLE "banner_urls";`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DROP TABLE "site_urls";`); err != nil {
		return err
	}

	_, err := tx.Exec(`DROP TABLE "banners";`)

	return err
}
