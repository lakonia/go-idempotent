package idempotency

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

const (
	defaultExpiry = 60 * time.Second
	defaultPrefix = "I"
)

var (
	ErrKeyAlreadyExists = errors.New("Key Exists")
)

type state struct {
	client *redis.Client
	prefix string
	expiry time.Duration
}

type configFn func(*state)

type Instance interface {
	CheckAndSet(ctx context.Context, idemKey string) error
	DeleteIdempotencyKey(ctx context.Context, idemKey string) error
}

func NewInstance(client *redis.Client, configFns ...configFn) Instance {
	s := &state{
		client: client,
		expiry: defaultExpiry,
		prefix: defaultPrefix,
	}

	for _, fn := range configFns {
		fn(s)
	}

	return s
}

func NoPrefix() configFn {
	return func(state *state) {
		state.prefix = ""
	}
}

func SetPrefix(prefix string) configFn {
	return func(state *state) {
		state.prefix = prefix
	}
}

func SetExpiry(expiry time.Duration) configFn {
	return func(state *state) {
		state.expiry = expiry
	}
}

func (str *state) DeleteIdempotencyKey(ctx context.Context, idemKey string) error {
	key := str.getRedisKey(idemKey)
	return str.client.WithContext(ctx).Del(ctx, key).Err()
}

func (str *state) CheckAndSet(ctx context.Context, idemKey string) error {
	key := str.getRedisKey(idemKey)

	res, err := str.client.WithContext(ctx).SetNX(ctx, key, 1, str.expiry).Result()
	if err != nil {
		return errors.Wrap(err, "Error while connecting to redis")
	}

	if res == false {
		return ErrKeyAlreadyExists
	}

	return nil
}

func (str *state) getRedisKey(idemKey string) string {
	if len(str.prefix) > 0 {
		return fmt.Sprintf("%s:%s", str.prefix, idemKey)
	}

	return idemKey
}
