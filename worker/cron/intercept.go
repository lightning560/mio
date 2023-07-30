package cron

import (
	"sync"
	"time"

	"miopkg/log"

	"github.com/robfig/cron/v3"
)

// delayIfStillRunning serializes jobs, delaying subsequent runs until the
// previous one is complete. Jobs running after a delay of more than a minute
// have the delay logged at Info.
func delayIfStillRunning(logger *log.Logger) JobWrapper {
	return func(j Job) Job {
		var mu sync.Mutex
		return cron.FuncJob(func() {
			start := time.Now()
			mu.Lock()
			defer mu.Unlock()
			if dur := time.Since(start); dur > time.Minute {
				logger.Info("cron delay", log.String("duration", dur.String()))
			}
			j.Run()
		})
	}
}

// skipIfStillRunning skips an invocation of the Job if a previous invocation is
// still running. It logs skips to the given logger at Info level.
func skipIfStillRunning(logger *log.Logger) JobWrapper {
	var ch = make(chan struct{}, 1)
	ch <- struct{}{}
	return func(j Job) Job {
		return cron.FuncJob(func() {
			select {
			case v := <-ch:
				j.Run()
				ch <- v
			default:
				logger.Info("cron skip")
			}
		})
	}
}
