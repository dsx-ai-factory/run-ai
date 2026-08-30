package keydb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/validator"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
)

const (
	maxKeysSafetyFactor = 2
)

type Validator interface {
	Validate(ctx context.Context) error
	Clear()
}

type Client struct {
	clientRW        *redis.Client
	clientR         *redis.Client
	storeTTLs       sync.Map
	storeValidators sync.Map
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Init() error {
	conf := config.Get[*cache_config.Cache]()

	clientRW, err := createClient(getKeyDBUrl(conf.Url))
	if err != nil {
		return fmt.Errorf("failed creating keydb RW client: %w", err)
	}

	var clientR *redis.Client
	if conf.ReadUrl == "" {
		clientR = clientRW
	} else if clientR, err = createClient(getKeyDBUrl(conf.ReadUrl)); err != nil {
		return fmt.Errorf("failed creating keydb R client: %w", err)
	}

	if err := checkKeyDBHealthy(clientRW, clientR); err != nil {
		return fmt.Errorf("failed to connect to keydb: %w", err)
	}

	c.clientRW = clientRW
	c.clientR = clientR
	return nil
}

func (c *Client) Store(cfg options.Store) error {
	if _, ok := c.storeValidators.Load(cfg.Name); ok {
		return nil
	}

	maxKeys := calculateStoreMaxKeys(cfg.Options)
	validator, err := validator.NewStoreLenValidator(cfg.Name, maxKeys, func(ctx context.Context) (int64, error) {
		return c.clientRW.HLen(ctx, cfg.Name).Result()
	})
	if err != nil {
		return err
	}

	c.storeTTLs.Store(cfg.Name, cfg.Options.Ttl)
	c.storeValidators.Store(cfg.Name, validator)

	return nil
}

func (c *Client) Close() error {
	if c.clientRW == c.clientR {
		return c.clientRW.Close()
	}

	var errs []error

	if err := c.clientRW.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close keydb RW client: %w", err))
	}

	if err := c.clientR.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close keydb R client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close keydb clients: %v", errs)
	}
	return nil
}

func (c *Client) SetIfAbsent(ctx context.Context, storeName string, key string, val []byte, ttl *time.Duration) (bool, error) {
	if err := c.validateStoreNotFull(ctx, storeName); err != nil {
		return false, err
	}

	acquired, err := c.clientRW.HSetNX(ctx, storeName, key, val).Result()
	if err != nil || !acquired {
		return acquired, err
	}
	ttlToUse := c.getKeyTTL(storeName)
	if ttl != nil {
		ttlToUse = *ttl
	}
	if ttlToUse == -1 {
		return true, nil
	}

	if err := c.clientRW.Do(ctx, "EXPIREMEMBER", storeName, key, int(ttlToUse.Seconds())).Err(); err != nil {
		return true, fmt.Errorf("failed to set expiration for key <%s,%s> ttl [%d] in cache: %w", storeName, key, int(ttlToUse.Seconds()), err)
	}

	return true, nil
}

func (c *Client) DeleteIf(ctx context.Context, storeName string, key string, cond func(val []byte) bool) error {
	deleteIfTransaction := func(tx *redis.Tx) error {
		val, err := tx.HGet(ctx, storeName, key).Result()
		if errors.Is(err, redis.Nil) || val == "" {
			return nil
		}
		if err != nil {
			return err
		}

		if !cond([]byte(val)) {
			return nil
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HDel(ctx, storeName, key)
			return nil
		})
		return err
	}
	for {
		if err := c.clientRW.Watch(ctx, deleteIfTransaction, storeName, key); errors.Is(err, redis.TxFailedErr) {
			continue
		} else if err != nil {
			return fmt.Errorf("failed to delete key <%s,%s> in cache: %w", storeName, key, err)
		}
		return nil
	}
}

func (c *Client) Set(ctx context.Context, storeName string, key string, val []byte) error {
	if err := c.validateStoreNotFull(ctx, storeName); err != nil {
		return err
	}

	if err := c.clientRW.HSet(ctx, storeName, key, val).Err(); err != nil {
		return fmt.Errorf("failed to set key <%s,%s> in cache: %w", storeName, key, err)
	}

	ttl := c.getKeyTTL(storeName)
	if ttl == -1 {
		return nil
	}

	if err := c.clientRW.Do(ctx, "EXPIREMEMBER", storeName, key, int(ttl.Seconds())).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for key <%s,%s> ttl [%d] in cache: %w", storeName, key, int(ttl.Seconds()), err)
	}

	return nil
}

func (c *Client) Get(ctx context.Context, storeName string, key string) ([]byte, error) {
	valBytes, err := c.clientR.HGet(ctx, storeName, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key <%s,%s> from cache: %w", storeName, key, err)
	}

	return valBytes, nil
}

func (c *Client) Delete(ctx context.Context, storeName string, key string) error {
	err := c.clientRW.HDel(ctx, storeName, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key <%s,%s> from cache: %w", storeName, key, err)
	}
	return nil
}

func (c *Client) ClearStore(ctx context.Context, storeName string) error {
	if err := c.clientRW.Del(ctx, storeName).Err(); err != nil {
		return fmt.Errorf("failed to clear store <%s>: %w", storeName, err)
	}

	v, ok := c.storeValidators.Load(storeName)
	if !ok {
		return fmt.Errorf("no validator found for store <%s> to nullify the counters", storeName)
	}

	v.(Validator).Clear()
	return nil

}

func (c *Client) getKeyTTL(storeName string) time.Duration {
	ttl, ok := c.storeTTLs.Load(storeName)
	if !ok {
		return -1 // = no TTL to set
	}
	ttlDuration, ok := ttl.(time.Duration)
	if !ok {
		return -1
	}
	return ttlDuration
}

func (c *Client) validateStoreNotFull(ctx context.Context, storeName string) error {
	v, ok := c.storeValidators.Load(storeName)
	if !ok {
		return fmt.Errorf("no validator found for store <%s>", storeName)
	}

	validator := v.(Validator)
	if err := validator.Validate(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Client) RetrieveMemory(storeName string) (int64, error) {
	return c.clientR.MemoryUsage(context.Background(), storeName, 0).Result()
}

func (c *Client) RetrieveLength(storeName string) (int64, error) {
	return c.clientR.HLen(context.Background(), storeName).Result()
}

func checkKeyDBHealthy(clientRW *redis.Client, clientR *redis.Client) error {
	ctx := context.Background()
	clientRW.Set(ctx, "Validation", "Redis connected successfully", time.Minute)
	_, err := clientR.Get(ctx, "Validation").Result()
	return err
}

func createClient(url string) (*redis.Client, error) {
	conf := config.Get[*cache_config.Cache]()

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("keydb URL parsing error: %w", err)
	}
	opt.Password = conf.Password
	opt.Username = conf.User

	opt.OnConnect = func(ctx context.Context, connection *redis.Conn) error {
		return connection.Ping(ctx).Err()
	}

	return redis.NewClient(opt), nil
}

func getKeyDBUrl(url string) string {
	return fmt.Sprintf("redis://%s", url)
}

func calculateStoreMaxKeys(opts *options.StoreOptions) int64 {
	maxKeys := (opts.MaxStoreSizeMB * 1000) / opts.MaxRecordSizeKB
	return int64(maxKeys * maxKeysSafetyFactor)
}
