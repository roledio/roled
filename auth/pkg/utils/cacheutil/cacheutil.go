package cacheutil

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/auth/pkg/utils/copyutil"
)

type Service interface {
	SetData(ctx context.Context, key string, val any, exp time.Duration) error
	GetData(ctx context.Context, key string, dest any) (bool, error)
}

type LoadFunc func(ctx context.Context) (any, error)

var (
	service Service
	once    sync.Once
)

func SetService(s Service) {
	if s == nil {
		panic("cacheutil: SetService called with nil")
	}
	once.Do(func() {
		service = s
	})
}

func isEmptySliceOrArray(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if kind == reflect.Pointer {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
		kind = rv.Kind()
	}
	if kind != reflect.Slice && kind != reflect.Array {
		return false
	}
	return rv.Len() == 0
}

func GetOrLoad(ctx context.Context, key string, dest any, ttl time.Duration, loader LoadFunc) error {
	if service == nil {
		return errors.New("cacheutil: service not set")
	}
	if key == "" {
		return errors.New("cacheutil: key is empty")
	}
	found, err := service.GetData(ctx, key, dest)
	if err != nil {
		log.WithContext(ctx).Warnw("cacheutil: failed to get data from cache, fallback to loader",
			"error", err, "key", key)
	} else if found {
		return nil
	}
	result, err := loader(ctx)
	if err != nil {
		return err
	}
	if isEmptySliceOrArray(result) {
		// Empty slice or array, no need to copy to dest or set to cache
		return nil
	}
	if err := copyutil.Copy(result, dest); err != nil {
		log.WithContext(ctx).Errorw("cacheutil: failed to copy loaded data to dest",
			"error", err, "key", key)
		return err
	}
	if err := service.SetData(ctx, key, result, ttl); err != nil {
		log.WithContext(ctx).Warnw("cacheutil: failed to set data to cache",
			"error", err, "key", key)
	}
	return nil
}

func SetMultipleKeys(ctx context.Context, keys []string, val any, ttl time.Duration) error {
	if service == nil {
		return errors.New("cacheutil: service not set")
	}
	for _, key := range keys {
		if key == "" {
			continue
		}
		if err := service.SetData(ctx, key, val, ttl); err != nil {
			log.WithContext(ctx).Warnw("cacheutil: failed to set data to cache",
				"error", err, "key", key)
		}
	}
	return nil
}
