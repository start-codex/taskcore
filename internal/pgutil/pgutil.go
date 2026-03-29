// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package pgutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// IsUniqueViolation reports whether err is a PostgreSQL unique-constraint violation (code 23505).
func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// WithTx begins a transaction, runs fn, and commits on success.
// defer tx.Rollback() is registered immediately after Begin so it fires on any return path.
// Begin errors are wrapped with beginLabel; Commit errors are wrapped with commitLabel.
// Errors returned by fn are returned as-is — each domain adds its own context inside fn.
func WithTx(ctx context.Context, db *sqlx.DB, opts *sql.TxOptions, beginLabel, commitLabel string, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("%s: %w", beginLabel, err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", commitLabel, err)
	}
	return nil
}
