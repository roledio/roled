package singleflightutil_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/roledio/roled/auth/pkg/utils/singleflightutil"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
)

func TestDo_Deduplication(t *testing.T) {
	var g singleflight.Group
	var calls int32
	const numGoroutines = 10

	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errs := make([]error, numGoroutines)
	shareds := make([]bool, numGoroutines)

	startGate := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-startGate
			res, err, shared := singleflightutil.Do(&g, "test-key", func() (string, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(50 * time.Millisecond)
				return "singleflight-result", nil
			})
			results[idx] = res
			errs[idx] = err
			shareds[idx] = shared
		}(i)
	}

	close(startGate)
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "Expected underlying func to be called only once")
	for i := 0; i < numGoroutines; i++ {
		assert.NoError(t, errs[i])
		assert.Equal(t, "singleflight-result", results[i])
	}
}

func TestDo_ErrorHandling(t *testing.T) {
	var g singleflight.Group
	expectedErr := errors.New("something went wrong")

	res, err, shared := singleflightutil.Do(&g, "error-key", func() (string, error) {
		return "", expectedErr
	})

	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, res)
	assert.False(t, shared)
}

func TestDo_NilGroupFallback(t *testing.T) {
	res, err, shared := singleflightutil.Do[string](nil, "nil-key", func() (string, error) {
		return "fallback-result", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "fallback-result", res)
	assert.False(t, shared)
}

func TestDoChan(t *testing.T) {
	var g singleflight.Group
	ch := singleflightutil.DoChan(&g, "chan-key", func() (int, error) {
		return 42, nil
	})

	res := <-ch
	assert.NoError(t, res.Err)
	assert.Equal(t, 42, res.Val)
}
