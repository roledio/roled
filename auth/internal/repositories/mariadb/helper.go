package mariadb

import (
	"context"
	"fmt"

	"github.com/roledio/roled/internal/repositories/interfaces"
)

var (
	ErrAffectedGreaterThanOne = fmt.Errorf("affected rows greater than one")
)

// exec executes a query and returns the number of rows affected.
func exec(ctx context.Context, qx interfaces.QueryExecutor, query string, args ...any) (int, error) {
	result, err := qx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

// execOne executes an update query and ensures that at most one row is affected.
// This is useful for operations like update or delete by ID, where we expect only one record to be modified.
//
// If more than one row is affected, it indicates a potential issue with the query or data integrity,
// as we expected to affect at most one row, and should be handled as a system error.
//
// If no rows are affected, it can be handled by the caller as a not found error, since it means the target record was not found.
func execOne(ctx context.Context, qx interfaces.QueryExecutor, query string, args ...any) (int, error) {
	rowsAffected, err := exec(ctx, qx, query, args...)
	if err != nil {
		return 0, err
	}
	if rowsAffected > 1 {
		return 0, ErrAffectedGreaterThanOne
	}
	return rowsAffected, nil
}

// namedExec executes a named query and returns the number of rows affected.
func namedExec(ctx context.Context, qx interfaces.QueryExecutor, query string, arg any) (int, error) {
	result, err := qx.NamedExecContext(ctx, query, arg)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rowsAffected), nil
}

// namedExecOne executes a named update query and ensures that at most one row is affected.
// This is useful for operations like update or delete by ID, where we expect only one record to be modified.
//
// If more than one row is affected, it indicates a potential issue with the query or data integrity,
// as we expected to affect at most one row, and should be handled as a system error.
//
// If no rows are affected, it can be handled by the caller as a not found error, since it means the target record was not found.
func namedExecOne(ctx context.Context, qx interfaces.QueryExecutor, query string, arg any) (int, error) {
	rowsAffected, err := namedExec(ctx, qx, query, arg)
	if err != nil {
		return 0, err
	}
	if rowsAffected > 1 {
		return 0, ErrAffectedGreaterThanOne
	}
	return rowsAffected, nil
}
