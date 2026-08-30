package repositories

import (
	"testing"

	"github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestToSQLOrderValue_Defaults(t *testing.T) {
	pr := models.PageRequest{}
	val := ToSQLOrderValue(pr, "name asc", "updated_at desc")
	assert.Equal(t, "name asc, updated_at desc", val)
}

func TestToSQLOrderValue_Specified(t *testing.T) {
	pr := models.PageRequest{SortBy: "name", SortDir: "desc"}
	val := ToSQLOrderValue(pr)
	assert.Equal(t, "name desc", val)
}

func TestWithAlias(t *testing.T) {
	assert.Equal(t, "u.id", WithAlias("id", "u"))
	assert.Equal(t, "name", WithAlias("name", ""))
}

func TestWithPercentHelpers(t *testing.T) {
	assert.Equal(t, "%a%", WithPercentAround("a"))
	assert.Equal(t, "%a", WithPercentBefore("a"))
	assert.Equal(t, "a%", WithPercentAfter("a"))
}
