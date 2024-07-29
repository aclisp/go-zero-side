package goschedule

import "time"

type SynchronizedClock struct {
	key                    string
	synchronizationManager SynchronizationManager
}

func NewSynchronizedClock(id string, syncMgr SynchronizationManager) *SynchronizedClock {
	clock := new(SynchronizedClock)
	clock.key = "clock:" + id
	clock.synchronizationManager = syncMgr
	return clock
}

func (clock *SynchronizedClock) Set(timestamp time.Time) bool {
	return clock.synchronizationManager.SetGreaterThan(clock.key, timestamp.UnixMilli())
}

func (clock *SynchronizedClock) Reset() {
	clock.synchronizationManager.Delete(clock.key)
}
