package goschedule

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

type ScheduledJob struct {
	job   gocron.Job
	clock *SynchronizedClock
}

var scheduler gocron.Scheduler
var synchronizationManager SynchronizationManager = NewSynchronizationManagerMemory()

func init() {
	var err error
	scheduler, err = gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	scheduler.Start()
}

func SetSynchronizationManager(syncMgr SynchronizationManager) {
	synchronizationManager = syncMgr
}

func ScheduleSynchronizedJob(id string, rule string, cb func()) (ScheduledJob, error) {
	clock := NewSynchronizedClock(id+":"+rule, synchronizationManager)

	var job gocron.Job
	var err error

	job, err = scheduler.NewJob(
		gocron.CronJob(rule, true),
		gocron.NewTask(func() {
			nextTimestamp, err := job.NextRun()
			if err != nil {
				logx.Errorf("Can not schedule synchronized job %q: job.NextRun: %v", id, err)
				return
			}

			wasSet := clock.Set(nextTimestamp)

			if wasSet {
				cb()
			}
		}),
	)
	if err != nil {
		return ScheduledJob{}, err
	}

	return ScheduledJob{job, clock}, nil
}

func (job ScheduledJob) Stop() {
	_ = scheduler.RemoveJob(job.job.ID())

	job.clock.Reset()
}
