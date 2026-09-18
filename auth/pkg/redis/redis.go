package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberredis "github.com/gofiber/storage/redis/v3"
	nrredis "github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Prefix   string
	DB       int
	Newrelic bool
}

// Service is a Redis storage service that extends fiber.Storage to provide additional methods for interacting with Redis.
type Service interface {
	fiber.Storage
	Ping() error
	SetData(ctx context.Context, key string, val any, exp time.Duration) error
	GetData(ctx context.Context, key string, dest any) (bool, error)
	KeysWithContext(ctx context.Context) ([]string, error)
	Keys() ([]string, error)
	Client() redis.UniversalClient
	KeyWithPrefix(key string) string
	DeleteManyWithContext(ctx context.Context, keys []string) error
}

type service struct {
	redisStorage *fiberredis.Storage
	prefix       string
}

func NewService(config *Config) Service {
	redisStorage := fiberredis.New(fiberredis.Config{
		Host:     config.Host,
		Port:     config.Port,
		Username: config.Username,
		Password: config.Password,
		Database: config.DB,
	})
	if config.Newrelic {
		redisStorage.Conn().AddHook(nrredis.NewHookWithOptions(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
			Username: config.Username,
			Password: config.Password,
			DB:       config.DB,
		}, nrredis.ConfigDatastoreKeysEnabled(true)))
	}
	return &service{
		redisStorage: redisStorage,
		prefix:       config.Prefix,
	}
}

func (p *service) KeyWithPrefix(key string) string {
	if key == "" || p.prefix == "" || strings.HasPrefix(key, p.prefix+":") {
		return key
	}
	return fmt.Sprintf("%s:%s", p.prefix, key)
}

func (p *service) Get(key string) ([]byte, error) {
	return p.redisStorage.Get(p.KeyWithPrefix(key))
}

func (p *service) GetWithContext(ctx context.Context, key string) ([]byte, error) {
	return p.redisStorage.GetWithContext(ctx, p.KeyWithPrefix(key))
}

func (p *service) GetData(ctx context.Context, key string, dest any) (bool, error) {
	data, err := p.GetWithContext(ctx, key)
	if err != nil {
		return false, err
	}
	// Key does not exist in redis
	if data == nil {
		return false, nil
	}
	err = json.Unmarshal(data, dest)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (p *service) Set(key string, val []byte, exp time.Duration) error {
	return p.redisStorage.Set(p.KeyWithPrefix(key), val, exp)
}

func (p *service) SetWithContext(ctx context.Context, key string, val []byte, exp time.Duration) error {
	return p.redisStorage.SetWithContext(ctx, p.KeyWithPrefix(key), val, exp)
}

func (p *service) SetData(ctx context.Context, key string, val any, exp time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return p.SetWithContext(ctx, key, data, exp)
}

// Deprecated: Use DeleteManyWithContext instead
func (p *service) Delete(key string) error {
	return p.redisStorage.Delete(p.KeyWithPrefix(key))
}

// Deprecated: Use DeleteManyWithContext instead
func (p *service) DeleteWithContext(ctx context.Context, key string) error {
	return p.redisStorage.DeleteWithContext(ctx, p.KeyWithPrefix(key))
}

func (p *service) DeleteManyWithContext(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = p.KeyWithPrefix(key)
	}
	// Use UNLINK (non-blocking delete) instead of DEL for better performance
	return p.redisStorage.Conn().Unlink(ctx, prefixedKeys...).Err()
}

func (p *service) Reset() error {
	return p.ResetWithContext(context.Background())
}

// ResetWithContext deletes all keys with the specified prefix. If no prefix is set,
// it flushes the entire database.
func (p *service) ResetWithContext(ctx context.Context) error {
	if p.prefix == "" {
		return p.redisStorage.ResetWithContext(ctx) // It will use FLUSHDB command
	}
	keys, err := p.KeysWithContext(ctx)
	if err != nil {
		return err
	}
	return p.DeleteManyWithContext(ctx, keys)
}

func (p *service) Close() error {
	return p.redisStorage.Close()
}

func (p *service) Keys() ([]string, error) {
	return p.KeysWithContext(context.Background())
}

func (p *service) KeysWithContext(ctx context.Context) ([]string, error) {
	keys := []string{}
	match := "*"
	if p.prefix != "" {
		match = fmt.Sprintf("%s:*", p.prefix)
	}
	scan := p.redisStorage.Conn().Scan(ctx, 0, match, 0)
	err := scan.Err()
	if err != nil {
		return nil, err
	}
	iter := scan.Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, nil
}

func (p *service) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.redisStorage.Conn().Ping(ctx).Err()
}

func (p *service) Client() redis.UniversalClient {
	return p.redisStorage.Conn()
}
