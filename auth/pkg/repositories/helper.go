package repositories

import (
	"strings"

	"github.com/roledio/roled/pkg/models"
)

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
