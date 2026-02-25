package workerpool

import "sync"

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
