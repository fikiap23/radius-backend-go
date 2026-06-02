package database

import (
	"context"
	"fmt"

	"github.com/radius/radius-backend/ent"
)

// RunInTx runs fn inside a single database transaction (BEGIN/COMMIT or ROLLBACK).
// Do not nest RunInTx calls; keep fn limited to fast DB work (no HTTP or long I/O).
func RunInTx(
	ctx context.Context,
	client *ent.Client,
	fn func(ctx context.Context, tx *ent.Client) error,
) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(ctx, tx.Client()); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rollback after error: %v (rollback: %w)", err, rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
