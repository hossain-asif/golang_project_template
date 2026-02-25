package scheduler

import (
	"context"
	"fmt"
	"time"
)

func TaskAssignment(ctx context.Context) {
	t := New()

	tasks := []Task{
		{
			Name:     "cleanup",
			Interval: 3 * time.Second,
			Fn: func(ctx context.Context) error {
				fmt.Println("running cleanup...")
				return nil
			},
		},
		{
			Name:     "healthcheck",
			Interval: 1 * time.Second,
			Fn: func(ctx context.Context) error {
				fmt.Println("running healthcheck...")
				return nil
			},
		},
	}

	t.StartAll(ctx, tasks)
}
