package xormcache

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"xorm.io/xorm/caches"
)

type RedisStore struct {
	store *redis.Redis
	Debug bool
}

var _ caches.CacheStore = (*RedisStore)(nil)

func NewRedisStore(redis *redis.Redis) *RedisStore {
	s := new(RedisStore)
	s.store = redis
	return s
}

func (s *RedisStore) Put(key string, value interface{}) error {
	val, err := caches.Encode(value)
	if err != nil {
		return err
	}

	return s.store.Set(key, string(val))
}

func (s *RedisStore) Get(key string) (interface{}, error) {
	val, err := s.store.Get(key)
	if err != nil {
		return nil, err
	}

	if val == "" {
		return nil, caches.ErrNotExist
	}

	var v interface{}
	err = caches.Decode([]byte(val), &v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *RedisStore) Del(key string) error {
	_, err := s.store.Del(key)
	return err
}
