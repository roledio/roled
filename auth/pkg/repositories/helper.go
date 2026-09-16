package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/roledio/roled/auth/pkg/models"
)

var (
	ErrAffectedGreaterThanOne = fmt.Errorf("affected rows greater than one")
)

// Exec executes a query and returns the number of rows affected.
func Exec(ctx context.Context, qx QueryExecutor, query string, args ...any) (int, error) {
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

// ExecOne executes an update query and ensures that at most one row is affected.
// This is useful for operations like update or delete by ID, where we expect only one record to be modified.
//
// If more than one row is affected, it indicates a potential issue with the query or data integrity,
// as we expected to affect at most one row, and should be handled as a system error.
//
// If no rows are affected, it can be handled by the caller as a not found error, since it means the target record was not found.
func ExecOne(ctx context.Context, qx QueryExecutor, query string, args ...any) (int, error) {
	rowsAffected, err := Exec(ctx, qx, query, args...)
	if err != nil {
		return 0, err
	}
	if rowsAffected > 1 {
		return 0, ErrAffectedGreaterThanOne
	}
	return rowsAffected, nil
}

// NamedExec executes a named query and returns the number of rows affected.
func NamedExec(ctx context.Context, qx QueryExecutor, query string, arg any) (int, error) {
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

// NamedExecOne executes a named update query and ensures that at most one row is affected.
// This is useful for operations like update or delete by ID, where we expect only one record to be modified.
//
// If more than one row is affected, it indicates a potential issue with the query or data integrity,
// as we expected to affect at most one row, and should be handled as a system error.
//
// If no rows are affected, it can be handled by the caller as a not found error, since it means the target record was not found.
func NamedExecOne(ctx context.Context, qx QueryExecutor, query string, arg any) (int, error) {
	rowsAffected, err := NamedExec(ctx, qx, query, arg)
	if err != nil {
		return 0, err
	}
	if rowsAffected > 1 {
		return 0, ErrAffectedGreaterThanOne
	}
	return rowsAffected, nil
}

// ToSQLOrderValue builds an SQL order value if the sort by and sort direction are provided.
// If not, it will build and return the default order value.
// Specify the default order value such as:
//
//	repository.ToSQLOrderValue(pageRequest, "name asc", "updated_at desc")
func ToSQLOrderValue(p models.PageRequest, defaultOrders ...string) string {
	if p.SortBy != "" && p.SortDir != "" {
		return p.SortBy + " " + p.SortDir
	}
	if len(defaultOrders) > 0 {
		return strings.Join(defaultOrders, ", ")
	}
	return ""
}

// FixSortDir normalizes the sort direction to either "ASC" or "DESC".
// If the input is empty or not "DESC", it defaults to "ASC".
func FixSortDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if !strings.EqualFold(dir, "desc") {
		return "ASC"
	}
	return "DESC"
}

func WithAlias(column, alias string) string {
	if alias == "" {
		return column
	}
	return alias + "." + column
}

func WithPercentAround(val string) string {
	return "%" + val + "%"
}

func WithPercentBefore(val string) string {
	return "%" + val
}

func WithPercentAfter(val string) string {
	return val + "%"
}
