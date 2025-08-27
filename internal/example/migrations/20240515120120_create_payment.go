package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreatePayment, downCreatePayment)
}

func upCreatePayment(ctx context.Context, txn *sql.Tx) error {
	_, err := txn.ExecContext(
		ctx,
		`
         CREATE TABLE payment
             (
                 id                  CHAR(14)     NOT NULL PRIMARY KEY,
                 created_at          BIGINT       NOT NULL,
                 updated_at          BIGINT       NOT NULL,
                 deleted_at          BIGINT       NULL,
                 amount              BIGINT       NOT NULL,
                 currency            VARCHAR(255) NOT NULL,
                 reference_id        VARCHAR(255) NOT NULL,
                 status              VARCHAR(55)  NOT NULL,
                 description         TEXT         NOT NULL,
                 error_code          VARCHAR(255),
                 error_message       VARCHAR(255),
                 payer               JSON         NOT NULL,
                 payee               JSON         NOT NULL
             );
     `)
	if err != nil {
		return err
	}
	_, err = txn.ExecContext(
		ctx,
		`CREATE INDEX idx_payment_reference_id on payment(reference_id)`)
	if err != nil {
		return err
	}
	return nil
}

func downCreatePayment(ctx context.Context, txn *sql.Tx) error {
	_, err := txn.Exec(`DROP INDEX idx_payment_reference_id ON payment;`)
	if err != nil {
		return err
	}
	_, err = txn.Exec(`DROP TABLE IF EXISTS payment`)
	if err != nil {
		return err
	}
	return err
}
