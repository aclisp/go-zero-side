package xormcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"xorm.io/xorm/caches"
)

func TestRedisGet(t *testing.T) {
	r, err := redis.NewRedis(redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)

	s, err := r.Get("abcaaa123")
	assert.NoError(t, err)
	assert.Empty(t, s)
}

func TestRedisStorePut(t *testing.T) {
	r, err := redis.NewRedis(redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)

	store := NewRedisStore(r)
	assert.NotNil(t, store)

	var kvs = map[string]interface{}{
		"Name": "Jack",
		"Age":  9223372036854775807,
	}

	for k, v := range kvs {
		assert.NoError(t, store.Put(k, v))
	}
}

func TestRedisStore(t *testing.T) {
	r, err := redis.NewRedis(redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
	})
	assert.NoError(t, err)
	assert.NotNil(t, r)

	store := NewRedisStore(r)
	assert.NotNil(t, store)

	var kvs = map[string]interface{}{
		"Name": "Jack",
		"Age":  9223372036854775807,
	}

	for k, v := range kvs {
		assert.NoError(t, store.Put(k, v))
	}

	for k, v := range kvs {
		assert.NoError(t, store.Put(k, v))
	}

	for k, v := range kvs {
		val, err := store.Get(k)
		assert.NoError(t, err)
		assert.EqualValues(t, v, val)
	}

	for k := range kvs {
		err := store.Del(k)
		assert.NoError(t, err)
	}

	for k := range kvs {
		_, err := store.Get(k)
		assert.EqualValues(t, caches.ErrNotExist, err)
	}
}
