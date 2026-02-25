package scheduler

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type Task struct {
    Name     string
    Interval time.Duration
    Fn       func(ctx context.Context) error
}

type Ticker struct{}

func New() *Ticker {
    return &Ticker{}
}

func (t *Ticker) run(ctx context.Context, task Task) {
    ticker := time.NewTicker(task.Interval)
    defer ticker.Stop()

    fmt.Printf("[%s] started with interval %s\n", task.Name, task.Interval)

    var mu sync.Mutex
    running := false

    for {
        select {
        case <-ctx.Done():
            fmt.Printf("[%s] stopped\n", task.Name)
            return
        case <-ticker.C:
            mu.Lock()
            if running {
                fmt.Printf("[%s] previous run still in progress, skipping\n", task.Name)
                mu.Unlock()
                continue
            }
            running = true
            mu.Unlock()

            go func() {
                defer func() {
                    if r := recover(); r != nil {
                        fmt.Printf("[%s] panicked: %v\n", task.Name, r)
                    }
                    mu.Lock()
                    running = false
                    mu.Unlock()
                }()

                if err := task.Fn(ctx); err != nil {
                    fmt.Printf("[%s] error: %v\n", task.Name, err)
                }
            }()
        }
    }
}

func (t *Ticker) StartAll(ctx context.Context, tasks []Task) {
    var wg sync.WaitGroup

    for _, task := range tasks {
        wg.Add(1)
        go func(tk Task) {
            defer wg.Done()
            t.run(ctx, tk)
        }(task)
    }

    wg.Wait()
    fmt.Println("all tasks stopped")
}