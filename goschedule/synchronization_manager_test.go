package goschedule

import (
	"testing"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestSynchronizationManagerMemory(t *testing.T) {
	memory := NewSynchronizationManagerMemory()

	ok := memory.SetGreaterThan("key", 12345)
	if !ok {
		t.Fail()
	}

	ok = memory.SetGreaterThan("key", 12345)
	if ok {
		t.Fail()
	}

	ok = memory.SetGreaterThan("key", 12346)
	if !ok {
		t.Fail()
	}
}

func TestSynchronizationManagerRedis(t *testing.T) {
	redis := NewSynchronizationManagerRedis(
		redis.MustNewRedis(redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}),
		"TestSynchronizationManagerRedis",
	)
	redis.Delete("key")

	ok := redis.SetGreaterThan("key", 12345)
	if !ok {
		t.Fail()
	}

	ok = redis.SetGreaterThan("key", 12345)
	if ok {
		t.Fail()
	}

	ok = redis.SetGreaterThan("key", 12346)
	if !ok {
		t.Fail()
	}
}
