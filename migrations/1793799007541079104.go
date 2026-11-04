package migrations

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/laukpauk/helper"
)

func init() {
	Migrations[1793799007541079104] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1793799007541079104"
		// transactions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE transactions (
				id UUID NOT NULL PRIMARY KEY DEFAULT uuidv7()
				, company_id UUID NOT NULL REFERENCES companies (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, parent_id UUID REFERENCES transactions (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, reordered BOOLEAN NOT NULL
				, user_id UUID NOT NULL REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, cart_id UUID NOT NULL REFERENCES carts (id) ON UPDATE CASCADE ON DELETE RESTRICT UNIQUE
				, reference_number character varying NOT NULL
				, status smallint NOT NULL
				, recipient_name character varying NOT NULL
				, recipient_mobile_phone character varying NOT NULL
				, buyer_note character varying NOT NULL
				, unique_code bigint NOT NULL
				, shopping_cost bigint NOT NULL
				, referrer_id UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, referral_code character varying
				, referral_points bigint NOT NULL
				, voucher_id UUID REFERENCES vouchers (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, voucher_type smallint
				, voucher_code character varying
				, voucher_discount bigint NOT NULL
				, reward_points_discounted BOOLEAN NOT NULL
				, reward_points_discount bigint NOT NULL
				, wallet_discounted BOOLEAN NOT NULL
				, wallet_amount bigint NOT NULL
				, wallet_discount bigint NOT NULL
				, discounted_shopping_cost bigint NOT NULL
				, delivery_estimated_cost bigint NOT NULL
				, discounted_delivery_estimated_cost bigint NOT NULL
				, delivery_route_id UUID NOT NULL REFERENCES delivery_routes (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, delivery_latitude character varying NOT NULL
				, delivery_longitude character varying NOT NULL
				, delivery_distance_km bigint NOT NULL
				, delivery_address character varying NOT NULL
				, delivery_instruction character varying NOT NULL
				, delivery_datetime timestamp with time zone NOT NULL
				, payment_amount bigint NOT NULL
				, payment_due_date timestamp with time zone
				, payment_channel_id UUID REFERENCES payment_channels (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, payment_channel_name character varying
				, payment_request_id character varying UNIQUE
				, payment_token_id character varying UNIQUE
				, payment_virtual_account_number character varying
				, payment_virtual_account_expired_at timestamp with time zone
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
				, cancelled_at timestamp with time zone
				, payment_status character varying
				, payment_failure_code character varying
				, payment_captured_at timestamp with time zone
				, payment_id character varying UNIQUE
				, paid_at timestamp with time zone
				, withdrawal_requested_at timestamp with time zone
				, withdrawal_reason character varying
				, withdrawn_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, withdrawn_at timestamp with time zone
				, processed_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, processed_at timestamp with time zone
				, revised_shopping_cost bigint NOT NULL
				, revised_discounted_shopping_cost bigint NOT NULL
				, revised_payment_amount bigint NOT NULL
				, refund_amount bigint NOT NULL
				, refunded_at timestamp with time zone
				, refund_id character varying UNIQUE
				, shipped_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, shipped_at timestamp with time zone
				, delivery_actual_cost bigint NOT NULL
				, complained_at timestamp with time zone
				, complaint character varying
				, completed_by UUID REFERENCES users (id) ON UPDATE CASCADE ON DELETE RESTRICT
				, completed_at timestamp with time zone
			);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (parent_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (reordered);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (company_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (user_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX transactions_unique_reference_number_idx ON transactions (LOWER(reference_number));`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (status);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (referrer_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (voucher_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (delivery_route_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_due_date);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_channel_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_request_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_token_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (payment_id);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (created_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (withdrawn_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (processed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (shipped_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (shipped_at);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON transactions (completed_by);`,
		); err != nil {
			helper.Capture(ctx, slog.LevelError, err, ctxt, "ErrExec")
		}
		return
	}
}
