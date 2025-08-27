package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateEventTable, downCreateEventTable)
}

func upCreateEventTable(ctx context.Context, txn *sql.Tx) error {
	_, err := txn.ExecContext(
		ctx,
		`
		CREATE TABLE event (
			id          character(14)     NOT NULL,
			data  		blob              NOT NULL,
			topic       character(50)     NOT NULL,
			created_at  integer           NOT NULL,
			updated_at  integer           NOT NULL,
			deleted_at  integer           DEFAULT NULL,
			PRIMARY KEY (id)
		);
	`)
	return err
}

func downCreateEventTable(ctx context.Context, txn *sql.Tx) error {
	_, err := txn.Exec(`DROP TABLE IF EXISTS event;`)
	return err
}
