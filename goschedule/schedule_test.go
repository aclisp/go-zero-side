package goschedule

import (
	"log"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func TestSchedule(t *testing.T) {
	SetSynchronizationManager(NewSynchronizationManagerRedis(
		redis.MustNewRedis(redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}),
		"TestSchedule",
	))

	job, err := ScheduleSynchronizedJob("test-schedule", "* * * * * *", func() {
		log.Println("xxxxxxxxx")
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)
	job.Stop()
}
