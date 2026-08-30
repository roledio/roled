package cacheutil

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	mocks "github.com/roledio/roled/auth/pkg/mocks/utils/cacheutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func resetService() {
	service = nil
	once = sync.Once{}
}

func TestIsEmptySliceOrArray(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"nil value", nil, false},
		{"empty string", "", false},
		{"string value", "hello", false},
		{"int value", 42, false},
		{"nil slice", ([]string)(nil), true},
		{"empty string slice", []string{}, true},
		{"non-empty string slice", []string{"a", "b"}, false},
		{"empty int slice", []int{}, true},
		{"non-empty int slice", []int{1, 2, 3}, false},
		{"empty struct slice", []struct{}{}, true},
		{"non-empty struct slice", []struct{ V int }{{V: 1}}, false},
		{"empty array", [0]string{}, true},
		{"non-empty array", [3]int{1, 2, 3}, false},
		{"pointer to empty slice", &[]string{}, true},
		{"pointer to non-empty slice", &[]int{1}, false},
		{"nil pointer to slice", (*[]string)(nil), false},
		{"map (not slice/array)", map[string]int{}, false},
		{"struct (not slice/array)", struct{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmptySliceOrArray(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOrLoad_ServiceNotSet(t *testing.T) {
	defer resetService()
	ctx := context.Background()
	var dest string
	err := GetOrLoad(ctx, "key", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return "loaded", nil
	})
	require.Error(t, err)
	assert.Equal(t, "cacheutil: service not set", err.Error())
}

func TestGetOrLoad_EmptyKey(t *testing.T) {
	defer resetService()
	SetService(mocks.NewMockService(t))

	ctx := context.Background()
	var dest string
	err := GetOrLoad(ctx, "", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return "loaded", nil
	})
	require.Error(t, err)
	assert.Equal(t, "cacheutil: key is empty", err.Error())
}

func TestGetOrLoad_CacheHit(t *testing.T) {
	defer resetService()
	mockSvc := mocks.NewMockService(t)
	SetService(mockSvc)

	ctx := context.Background()
	var dest string
	loaderCalled := false

	mockSvc.On("GetData", ctx, "key1", &dest).Return(true, nil).Run(func(args mock.Arguments) {
		destPtr := args.Get(2).(*string)
		*destPtr = "cached_value"
	})

	err := GetOrLoad(ctx, "key1", &dest, time.Minute, func(ctx context.Context) (any, error) {
		loaderCalled = true
		return "loaded", nil
	})
	require.NoError(t, err)
	assert.False(t, loaderCalled)
	assert.Equal(t, "cached_value", dest, "dest should be populated from cache on hit")
	mockSvc.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_SetsCache(t *testing.T) {
	defer resetService()
	mockSvc := mocks.NewMockService(t)
	SetService(mockSvc)

	ctx := context.Background()
	var dest string

	mockSvc.On("GetData", ctx, "key1", &dest).Return(false, nil)
	mockSvc.On("SetData", ctx, "key1", "loaded_value", time.Minute).Return(nil)

	err := GetOrLoad(ctx, "key1", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return "loaded_value", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "loaded_value", dest, "dest should be populated from loader on miss")
	mockSvc.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_EmptySlice_NotCached(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest []string

	mock.On("GetData", ctx, "key_empty", &dest).Return(false, nil)

	err := GetOrLoad(ctx, "key_empty", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return []string{}, nil
	})
	require.NoError(t, err)
	assert.Nil(t, dest, "dest should remain nil (zero-value) for empty slice not cached nor copied")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_NilSlice_NotCached(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest []string

	mock.On("GetData", ctx, "key_nil", &dest).Return(false, nil)

	err := GetOrLoad(ctx, "key_nil", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return ([]string)(nil), nil
	})
	require.NoError(t, err)
	assert.Nil(t, dest, "dest should remain nil for nil slice not cached nor copied")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_NonEmptySlice_Cached(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest []string
	expected := []string{"a", "b", "c"}

	mock.On("GetData", ctx, "key_slice", &dest).Return(false, nil)
	mock.On("SetData", ctx, "key_slice", expected, time.Minute).Return(nil)

	err := GetOrLoad(ctx, "key_slice", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return expected, nil
	})
	require.NoError(t, err)
	assert.Equal(t, expected, dest, "dest should be populated from loader with non-empty slice")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_EmptyArray_NotCached(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest [0]int

	mock.On("GetData", ctx, "key_arr_empty", &dest).Return(false, nil)

	err := GetOrLoad(ctx, "key_arr_empty", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return [0]int{}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, [0]int{}, dest, "dest should be zero-value empty array for empty array result")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_NonEmptyArray_Cached(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest [2]string
	expected := [2]string{"x", "y"}

	mock.On("GetData", ctx, "key_arr", &dest).Return(false, nil)
	mock.On("SetData", ctx, "key_arr", expected, time.Minute).Return(nil)

	err := GetOrLoad(ctx, "key_arr", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return expected, nil
	})
	require.NoError(t, err)
	assert.Equal(t, expected, dest, "dest should be populated from loader with non-empty array")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_Struct_Cached(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	type testStruct struct {
		ID   int
		Name string
	}

	ctx := context.Background()
	var dest testStruct
	expected := testStruct{ID: 1, Name: "test"}

	mock.On("GetData", ctx, "key_struct", &dest).Return(false, nil)
	mock.On("SetData", ctx, "key_struct", expected, time.Minute).Return(nil)

	err := GetOrLoad(ctx, "key_struct", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return expected, nil
	})
	require.NoError(t, err)
	assert.Equal(t, expected, dest, "dest should be populated from loader with struct")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_CacheMiss_LoaderError(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest string
	expectedErr := errors.New("load failed")

	mock.On("GetData", ctx, "key_err", &dest).Return(false, nil)

	err := GetOrLoad(ctx, "key_err", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return nil, expectedErr
	})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", dest, "dest should remain zero-value on loader error")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_GetDataError_FallsBackToLoader(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest string

	mock.On("GetData", ctx, "key_fallback", &dest).Return(false, errors.New("redis down"))
	mock.On("SetData", ctx, "key_fallback", "from_loader", time.Minute).Return(nil)

	err := GetOrLoad(ctx, "key_fallback", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return "from_loader", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "from_loader", dest, "dest should be populated from loader after GetData error fallback")
	mock.AssertExpectations(t)
}

func TestGetOrLoad_SetDataError_StillSucceeds(t *testing.T) {
	defer resetService()
	mock := mocks.NewMockService(t)
	SetService(mock)

	ctx := context.Background()
	var dest string

	mock.On("GetData", ctx, "key_seterr", &dest).Return(false, nil)
	mock.On("SetData", ctx, "key_seterr", "loaded", time.Minute).Return(errors.New("set failed"))

	err := GetOrLoad(ctx, "key_seterr", &dest, time.Minute, func(ctx context.Context) (any, error) {
		return "loaded", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "loaded", dest, "dest should be populated even if SetData fails")
	mock.AssertExpectations(t)
}
