package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/roledio/roled/auth/pkg/models"
	"github.com/roledio/roled/auth/pkg/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockResult struct {
	mock.Mock
}

func (m *mockResult) LastInsertId() (int64, error) {
	mockArgs := m.Called()
	return mockArgs.Get(0).(int64), mockArgs.Error(1)
}

func (m *mockResult) RowsAffected() (int64, error) {
	mockArgs := m.Called()
	return mockArgs.Get(0).(int64), mockArgs.Error(1)
}

func TestExec(t *testing.T) {
	ctx := context.Background()

	t.Run("executes query successfully", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1", mock.Anything).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(5), nil)

		rowsAffected, err := Exec(ctx, mockQx, "UPDATE test SET x = 1")

		assert.NoError(t, err)
		assert.Equal(t, 5, rowsAffected)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("returns error on query execution failure", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1", mock.Anything).Return(nil, errors.New("db error"))

		rowsAffected, err := Exec(ctx, mockQx, "UPDATE test SET x = 1")

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		assert.Contains(t, err.Error(), "db error")
		mockQx.AssertExpectations(t)
	})

	t.Run("returns error on rows affected failure", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1", mock.Anything).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(0), errors.New("rows error"))

		rowsAffected, err := Exec(ctx, mockQx, "UPDATE test SET x = 1")

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		assert.Contains(t, err.Error(), "rows error")
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})
}

func TestExecOne(t *testing.T) {
	ctx := context.Background()

	t.Run("executes one row successfully", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1 WHERE id = ?", mock.Anything).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(1), nil)

		rowsAffected, err := ExecOne(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = ?", "abc")

		assert.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("returns error when more than one row affected", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1", mock.Anything).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(2), nil)

		rowsAffected, err := ExecOne(ctx, mockQx, "UPDATE test SET x = 1")

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		assert.Equal(t, ErrAffectedGreaterThanOne, err)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("returns zero rows when no rows affected", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1 WHERE id = ?", mock.Anything).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(0), nil)

		rowsAffected, err := ExecOne(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = ?", "nonexistent")

		assert.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("propagates query execution error", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)

		mockQx.On("ExecContext", ctx, "UPDATE test SET x = 1", mock.Anything).Return(nil, errors.New("db error"))

		rowsAffected, err := ExecOne(ctx, mockQx, "UPDATE test SET x = 1")

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		mockQx.AssertExpectations(t)
	})
}

func TestNamedExec(t *testing.T) {
	ctx := context.Background()

	t.Run("executes named query successfully", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		arg := map[string]any{"id": "abc", "value": 1}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = :value WHERE id = :id", arg).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(3), nil)

		rowsAffected, err := NamedExec(ctx, mockQx, "UPDATE test SET x = :value WHERE id = :id", arg)

		assert.NoError(t, err)
		assert.Equal(t, 3, rowsAffected)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("returns error on named execution failure", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)

		arg := map[string]any{"id": "abc"}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = 1 WHERE id = :id", arg).Return(nil, errors.New("named error"))

		rowsAffected, err := NamedExec(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = :id", arg)

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		assert.Contains(t, err.Error(), "named error")
		mockQx.AssertExpectations(t)
	})

	t.Run("returns error on rows affected failure", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		arg := map[string]any{"id": "abc"}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = 1 WHERE id = :id", arg).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(0), errors.New("rows error"))

		rowsAffected, err := NamedExec(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = :id", arg)

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		assert.Contains(t, err.Error(), "rows error")
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})
}

func TestNamedExecOne(t *testing.T) {
	ctx := context.Background()

	t.Run("executes named query for one row successfully", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		arg := map[string]any{"id": "abc", "value": 1}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = :value WHERE id = :id", arg).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(1), nil)

		rowsAffected, err := NamedExecOne(ctx, mockQx, "UPDATE test SET x = :value WHERE id = :id", arg)

		assert.NoError(t, err)
		assert.Equal(t, 1, rowsAffected)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("returns error when more than one row affected", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		arg := map[string]any{"id": "abc"}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = 1 WHERE id = :id", arg).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(5), nil)

		rowsAffected, err := NamedExecOne(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = :id", arg)

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		assert.Equal(t, ErrAffectedGreaterThanOne, err)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("returns zero rows when no rows affected", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)
		mockRes := new(mockResult)

		arg := map[string]any{"id": "nonexistent"}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = 1 WHERE id = :id", arg).Return(mockRes, nil)
		mockRes.On("RowsAffected").Return(int64(0), nil)

		rowsAffected, err := NamedExecOne(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = :id", arg)

		assert.NoError(t, err)
		assert.Equal(t, 0, rowsAffected)
		mockQx.AssertExpectations(t)
		mockRes.AssertExpectations(t)
	})

	t.Run("propagates named execution error", func(t *testing.T) {
		mockQx := mocks.NewMockQueryExecutor(t)

		arg := map[string]any{"id": "abc"}
		mockQx.On("NamedExecContext", ctx, "UPDATE test SET x = 1 WHERE id = :id", arg).Return(nil, errors.New("named error"))

		rowsAffected, err := NamedExecOne(ctx, mockQx, "UPDATE test SET x = 1 WHERE id = :id", arg)

		assert.Error(t, err)
		assert.Equal(t, 0, rowsAffected)
		mockQx.AssertExpectations(t)
	})
}

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

func TestToSQLOrderValue_EmptyDefaults(t *testing.T) {
	pr := models.PageRequest{}
	val := ToSQLOrderValue(pr)
	assert.Equal(t, "", val)
}

func TestFixSortDir(t *testing.T) {
	t.Run("returns DESC for desc (case-insensitive)", func(t *testing.T) {
		assert.Equal(t, "DESC", FixSortDir("desc"))
		assert.Equal(t, "DESC", FixSortDir("DESC"))
		assert.Equal(t, "DESC", FixSortDir("Desc"))
		assert.Equal(t, "DESC", FixSortDir("dEsC"))
	})

	t.Run("returns ASC for empty string", func(t *testing.T) {
		assert.Equal(t, "ASC", FixSortDir(""))
	})

	t.Run("returns ASC for whitespace", func(t *testing.T) {
		assert.Equal(t, "ASC", FixSortDir("  "))
	})

	t.Run("returns ASC for invalid values", func(t *testing.T) {
		assert.Equal(t, "ASC", FixSortDir("asc"))
		assert.Equal(t, "ASC", FixSortDir("invalid"))
		assert.Equal(t, "ASC", FixSortDir("ascending"))
		assert.Equal(t, "ASC", FixSortDir("random"))
	})

	t.Run("trims whitespace before checking", func(t *testing.T) {
		assert.Equal(t, "DESC", FixSortDir("  desc  "))
		assert.Equal(t, "ASC", FixSortDir("  asc  "))
	})
}

func TestWithAlias(t *testing.T) {
	assert.Equal(t, "u.id", WithAlias("id", "u"))
	assert.Equal(t, "name", WithAlias("name", ""))
	assert.Equal(t, "users.email", WithAlias("email", "users"))
}

func TestWithPercentHelpers(t *testing.T) {
	assert.Equal(t, "%a%", WithPercentAround("a"))
	assert.Equal(t, "%a", WithPercentBefore("a"))
	assert.Equal(t, "a%", WithPercentAfter("a"))
}
