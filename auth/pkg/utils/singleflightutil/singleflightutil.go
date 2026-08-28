package singleflightutil

import (
	"golang.org/x/sync/singleflight"
)

// Group is an alias for singleflight.Group.
type Group = singleflight.Group

// Do executes and returns the results of the given function, making sure that
// only one execution is in-flight for a given key at any point in time.
// If a duplicate request arrives while another request is in-flight for key,
// the duplicate caller waits for the original to complete and receives the exact same results.
func Do[T any](g *singleflight.Group, key string, fn func() (T, error)) (T, error, bool) {
	if g == nil {
		v, err := fn()
		return v, err, false
	}
	v, err, shared := g.Do(key, func() (any, error) {
		return fn()
	})
	if v == nil {
		var zero T
		return zero, err, shared
	}
	return v.(T), err, shared
}

// DoChan executes and returns a channel that will receive the results when available.
func DoChan[T any](g *singleflight.Group, key string, fn func() (T, error)) <-chan singleflight.Result {
	if g == nil {
		ch := make(chan singleflight.Result, 1)
		v, err := fn()
		ch <- singleflight.Result{Val: v, Err: err, Shared: false}
		close(ch)
		return ch
	}
	return g.DoChan(key, func() (any, error) {
		return fn()
	})
}
