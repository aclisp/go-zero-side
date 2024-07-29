package goschedule

import (
	_ "embed"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type SynchronizationManager interface {
	Set(key string, value int64)
	Get(key string) (value int64, ok bool)
	Delete(key string)
	Exists(key string) bool
	SetGreaterThan(key string, value int64) bool
}

var _ SynchronizationManager = (*synchronizationManagerMemory)(nil) // Verify that *T implements I.
var _ SynchronizationManager = (*synchronizationManagerRedis)(nil)

//go:embed set_greater_than.lua
var setGreaterThanLua string
var setGreaterThanScript = redis.NewScript(setGreaterThanLua)

type synchronizationManagerMemory struct {
	store map[string]int64
}

func NewSynchronizationManagerMemory() SynchronizationManager {
	memory := new(synchronizationManagerMemory)
	memory.store = make(map[string]int64)
	return memory
}

func (memory *synchronizationManagerMemory) Set(key string, value int64) {
	memory.store[key] = value
}

func (memory *synchronizationManagerMemory) Get(key string) (value int64, ok bool) {
	value, ok = memory.store[key]
	return
}

func (memory *synchronizationManagerMemory) Delete(key string) {
	delete(memory.store, key)
}

func (memory *synchronizationManagerMemory) Exists(key string) (ok bool) {
	_, ok = memory.store[key]
	return
}

func (memory *synchronizationManagerMemory) SetGreaterThan(key string, value int64) bool {
	if memory.Exists(key) {
		oldValue, _ := memory.Get(key)

		if value <= oldValue {
			return false
		}
	}

	memory.Set(key, value)
	return true
}

type synchronizationManagerRedis struct {
	namespace string
	store     *redis.Redis
}

func NewSynchronizationManagerRedis(store *redis.Redis, namespace string) SynchronizationManager {
	s := new(synchronizationManagerRedis)
	s.store = store
	s.namespace = namespace
	return s
}

func (s *synchronizationManagerRedis) getNamespacedKey(key string) string {
	return s.namespace + ":" + key
}

func (s *synchronizationManagerRedis) Set(key string, value int64) {
	err := s.store.Set(s.getNamespacedKey(key), strconv.FormatInt(value, 10))
	logx.Must(err)
}

func (s *synchronizationManagerRedis) Get(key string) (value int64, ok bool) {
	str, err := s.store.Get(s.getNamespacedKey(key))
	if err != nil {
		return 0, false
	}
	value, _ = strconv.ParseInt(str, 10, 64)
	ok = true
	return
}

func (s *synchronizationManagerRedis) Delete(key string) {
	_, _ = s.store.Del(s.getNamespacedKey(key))
}

func (s *synchronizationManagerRedis) Exists(key string) bool {
	ok, err := s.store.Exists(s.getNamespacedKey(key))
	if err != nil {
		return false
	}
	return ok
}

func (s *synchronizationManagerRedis) SetGreaterThan(key string, value int64) bool {
	wasSet, err := s.store.ScriptRun(setGreaterThanScript, []string{s.getNamespacedKey(key)}, value)
	if err != nil {
		return false
	}
	if n, ok := wasSet.(int64); ok && n == 1 {
		return true
	}
	return false
}
