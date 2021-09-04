package idempotency

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var (
	ErrKeyExists = errors.New("Key Exists")
)

type state struct {
	client *redis.Client
	Prefix string
	Key    string
}

type Instance interface {
	GetKey() string
	CheckAndSet(ctx context.Context, idemKey string) error
	DeleteIdempotencyKey(ctx context.Context, idemKey string) error
}

func NewInstance(client *redis.Client, prefix string, key string) Instance {
	return &state{
		client: client,
		Prefix: prefix,
		Key:    key,
	}
}

func (str *state) GetKey() string {
	return str.Key
}

func (str *state) DeleteIdempotencyKey(ctx context.Context, idemKey string) error {
	key := getRedisKey(str.Prefix, idemKey)
	return str.client.Del(ctx, key).Err()
}

func (str *state) CheckAndSet(ctx context.Context, idemKey string) error {
	key := getRedisKey(str.Prefix, idemKey)
	scr := redis.NewScript(`
				if redis.call("EXISTS", KEYS[1]) == 1 then
					return "1"
				end
				redis.call("SET", KEYS[1], 1)
				redis.call("EXPIRE", KEYS[1], ARGV[1])
				return "0"
			`)

	c, err := scr.Run(ctx, str.client, []string{key}, 60).Result()
	if err != nil {
		return errors.Wrap(err, "Couldn't run the lua script")
	}

	if c.(string) == "1" {
		return ErrKeyExists
	}

	return nil
}

func getRedisKey(prefix, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}
