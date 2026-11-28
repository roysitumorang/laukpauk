package migrations

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799007541079104] = Migration{
		Version: 1793799007541079104,
		Up: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Up"
			// transactions
			if _, err = tx.ExecContext(
				ctx,
				`CREATE TABLE transactions (
					id UUID NOT NULL PRIMARY KEY DEFAULT (UUID_v7())
					, parent_id UUID
					, reordered BOOLEAN NOT NULL
					, user_id UUID NOT NULL
					, cart_id UUID NOT NULL
					, reference_number VARCHAR(255) NOT NULL
					, status smallint NOT NULL
					, recipient_name TEXT NOT NULL
					, recipient_mobile_phone TEXT NOT NULL
					, buyer_note TEXT NOT NULL
					, unique_code bigint NOT NULL
					, shopping_cost bigint NOT NULL
					, referrer_id UUID
					, referral_code TEXT
					, referral_points bigint NOT NULL
					, voucher_id UUID
					, voucher_type smallint
					, voucher_discount bigint NOT NULL
					, promotion_id UUID
					, promotion_code TEXT
					, promotion_discount bigint NOT NULL
					, reward_points_discounted BOOLEAN NOT NULL
					, reward_points_discount bigint NOT NULL
					, wallet_discounted BOOLEAN NOT NULL
					, wallet_amount bigint NOT NULL
					, wallet_discount bigint NOT NULL
					, discounted_shopping_cost bigint NOT NULL
					, delivery_estimated_cost bigint NOT NULL
					, discounted_delivery_estimated_cost bigint NOT NULL
					, delivery_address_id UUID NOT NULL
					, delivery_destination_latitude TEXT NOT NULL
					, delivery_destination_longitude TEXT NOT NULL
					, delivery_distance_km bigint NOT NULL
					, delivery_street_address TEXT NOT NULL
					, delivery_street_address_note TEXT NOT NULL
					, delivery_instruction TEXT NOT NULL
					, delivery_datetime DATETIME(6) NOT NULL
					, payment_amount bigint NOT NULL
					, payment_due_date DATETIME(6)
					, payment_channel_id UUID
					, payment_channel_name TEXT
					, payment_request_id VARCHAR(255)
					, payment_token_id VARCHAR(255)
					, payment_virtual_account_number TEXT
					, payment_virtual_account_expired_at DATETIME(6)
					, created_at DATETIME(6) NOT NULL
					, updated_at DATETIME(6) NOT NULL
					, cancelled_at DATETIME(6)
					, payment_status TEXT
					, payment_failure_code TEXT
					, payment_captured_at DATETIME(6)
					, payment_id VARCHAR(255)
					, paid_at DATETIME(6)
					, withdrawal_requested_at DATETIME(6)
					, withdrawal_reason TEXT
					, withdrawn_by UUID
					, withdrawn_at DATETIME(6)
					, processed_by UUID
					, processed_at DATETIME(6)
					, revised_shopping_cost bigint NOT NULL
					, revised_discounted_shopping_cost bigint NOT NULL
					, revised_payment_amount bigint NOT NULL
					, refund_amount bigint NOT NULL
					, refunded_at DATETIME(6)
					, refund_id VARCHAR(255)
					, shipped_by UUID
					, shipped_at DATETIME(6)
					, delivery_actual_cost bigint NOT NULL
					, complained_at DATETIME(6)
					, complaint TEXT
					, completed_by UUID
					, completed_at DATETIME(6)
					, reference_number_key VARCHAR(255) AS (LOWER(reference_number)) PERSISTENT INVISIBLE
					, CONSTRAINT transactions_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES carts (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_cart_id_key UNIQUE (cart_id)
					, CONSTRAINT transactions_referrer_id_fkey FOREIGN KEY (referrer_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_voucher_id_fkey FOREIGN KEY (voucher_id) REFERENCES vouchers (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_delivery_address_id_fkey FOREIGN KEY (delivery_address_id) REFERENCES delivery_addresses (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_payment_channel_id_fkey FOREIGN KEY (payment_channel_id) REFERENCES payment_channels (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_payment_request_id_key UNIQUE (payment_request_id)
					, CONSTRAINT transactions_payment_token_id_key UNIQUE (payment_token_id)
					, CONSTRAINT transactions_payment_id_key UNIQUE (payment_id)
					, CONSTRAINT transactions_withdrawn_by_fkey FOREIGN KEY (withdrawn_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_processed_by_fkey FOREIGN KEY (processed_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_refund_id_key UNIQUE (refund_id)
					, CONSTRAINT transactions_shipped_by_fkey FOREIGN KEY (shipped_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_completed_by_fkey FOREIGN KEY (completed_by) REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
					, CONSTRAINT transactions_unique_reference_number_idx UNIQUE (reference_number_key)
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_as_cs;`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_parent_id_idx ON transactions (parent_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_reordered_idx ON transactions (reordered);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_user_id_idx ON transactions (user_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_status_idx ON transactions (status);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_referrer_id_idx ON transactions (referrer_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_voucher_id_idx ON transactions (voucher_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_delivery_address_id_idx ON transactions (delivery_address_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_payment_due_date_idx ON transactions (payment_due_date);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_payment_channel_id_idx ON transactions (payment_channel_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_payment_request_id_idx ON transactions (payment_request_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_payment_token_id_idx ON transactions (payment_token_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_payment_id_idx ON transactions (payment_id);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_created_at_idx ON transactions (created_at);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_withdrawn_by_idx ON transactions (withdrawn_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_processed_by_idx ON transactions (processed_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_shipped_by_idx ON transactions (shipped_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_shipped_at_idx ON transactions (shipped_at);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
				return
			}
			if _, err = tx.ExecContext(
				ctx,
				`CREATE INDEX transactions_completed_by_idx ON transactions (completed_by);`,
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
		Down: func(ctx context.Context, tx *sql.Tx) (err error) {
			ctxt := "Migration-Down"
			if _, err = tx.ExecContext(
				ctx,
				"DROP TABLE transactions",
			); err != nil {
				helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExecContext")
			}
			return
		},
	}
}
