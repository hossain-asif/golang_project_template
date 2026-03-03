package scheduler

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"sync"
	"time"
)

// global decalaration
var schedulerLog = logger.Log.Scope("common_pkg", "scheduler", "ticker")

type Task struct {
	Name     string
	Interval time.Duration
	Fn       func(ctx context.Context) error
}

type Ticker struct {
}

func NewTicker() *Ticker {
	return &Ticker{}
}

/*
Run starts a goroutine that runs the given task function at the specified interval.
If the context is canceled, the goroutine will stop and print a message indicating that it was stopped.
If a previous run of the task function is still in progress when the next tick occurs, the goroutine will skip that run and print a message indicating that it was skipped.
If the task function panics, the goroutine will recover and print a message indicating that it panicked.
*/
func (t *Ticker) run(ctx context.Context, task Task) {
	log := schedulerLog.Method("run").WithContext(ctx)

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	log.Infof("%s started with interval %s\n", task.Name, task.Interval)

	var mu sync.Mutex
	running := false

	for {
		select {
		case <-ctx.Done():
			log.Infof("%s stopped\n", task.Name)
			return
		case <-ticker.C:
			mu.Lock()
			if running {
				log.Infof("[%s] skipping, previous run still in progress\n", task.Name)
				mu.Unlock()
				continue
			}
			running = true
			mu.Unlock()

			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf("[%s] panicked: %v\n", task.Name, r)
					}
					mu.Lock()
					running = false
					mu.Unlock()
				}()

				if err := task.Fn(ctx); err != nil {
					log.Errorf("[%s] error: %v\n", task.Name, err)
				}
			}()
		}
	}
}

func (t *Ticker) StartAll(ctx context.Context, tasks []Task) {
	log := schedulerLog.Method("StartAll").WithContext(ctx)
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(tk Task) {
			defer wg.Done()
			t.run(ctx, tk)
		}(task)
	}

	wg.Wait()
	log.Infof("all tasks stopped")
}
