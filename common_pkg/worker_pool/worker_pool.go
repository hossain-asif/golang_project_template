package workerpool

import "sync"

/*
 This is thread pool pattern with fixed number of workers. For go-lang it is called Worker Pool or Goroutine Pool.
 	- Thread Pool Pattern — fixed number of workers reused across jobs, no new goroutine per job
	- Fan-out — one producer distributes work across multiple workers
	- Fan-in — multiple workers funnel results into one result channel
	- Pipeline Pattern — data flows through stages: read → batch → process → collect
*/

type Job[T any] struct {
	Data T
}

type Result[T any] struct {
	Data T
	Err  error
}

type WorkerPool[I any, O any] struct {
	JobChan     chan Job[I]
	ResultChan  chan Result[O]
	Process     func(I) (O, error)
	WorkerCount int
	wg          sync.WaitGroup
}

func NewWorkerPool[I any, O any](workerCount int, bufferSize int, process func(I) (O, error)) *WorkerPool[I, O] {
	return &WorkerPool[I, O]{
		JobChan:     make(chan Job[I], bufferSize),
		ResultChan:  make(chan Result[O], bufferSize),
		Process:     process,
		WorkerCount: workerCount,
	}
}

// Start spins up workers
func (wp *WorkerPool[I, O]) Start() {
	for i := 0; i < wp.WorkerCount; i++ {
		wp.wg.Add(1)
		go wp.RunWorker()
	}

	go func() {
		wp.wg.Wait()
		close(wp.ResultChan)
	}()
}

func (wp *WorkerPool[I, O]) RunWorker() {
	defer wp.wg.Done()

	for job := range wp.JobChan {
		output, err := wp.Process(job.Data)
		wp.ResultChan <- Result[O]{Data: output, Err: err}
	}
}

func (wp *WorkerPool[I, O]) Submit(Data I) {
	wp.JobChan <- Job[I]{Data: Data}
}

func (wp *WorkerPool[I, O]) Done() {
	close(wp.JobChan)
}

func (wp *WorkerPool[I, O]) Results() <-chan Result[O] {
	return wp.ResultChan
}

func NewPool[I any](workerCount int, process func(I) error) *WorkerPool[I, struct{}] {
	pool := NewWorkerPool[I, struct{}](
		workerCount,
		workerCount,
		func(data I) (struct{}, error) {
			return struct{}{}, process(data)
		},
	)
	pool.Start()
	return pool
}
