package scheduler

import (
	"context"
)

func TaskAssignment(ctx context.Context, tasks []Task) {
	ticker := NewTicker()
	ticker.StartAll(ctx, tasks)
}
