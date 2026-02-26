package scheduler

import (
	"context"
)

func TaskAssignment(ctx context.Context, tasks []Task) {
	t := NewTicker()
	t.StartAll(ctx, tasks)
}
