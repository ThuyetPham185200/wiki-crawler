package workers

import (
	"time"
	"wikicrawler/internal/utils/processor"
	tasks "wikicrawler/internal/utils/workers/task"
)

type Worker struct {
	processor.BaseProcessor
	workerpool *WorkerPool
}

func NewWorker(wp *WorkerPool) *Worker {
	s := &Worker{
		workerpool: wp,
	}
	s.Init(s)
	return s
}

// RunningTask implement từ ServerProcessor
func (s *Worker) RunningTask() {
	task := s.workerpool.Tasks.TryPop(1 * time.Millisecond)
	if task != nil {
		task.DoTask()
	}
}

type WorkerPool struct {
	Tasks   *tasks.TaskQueue
	Workers []*Worker
}

func NewWorkerPool(numworkers int, ntasks int) *WorkerPool {
	return &WorkerPool{
		Tasks:   tasks.NewTaskQueue(ntasks),
		Workers: make([]*Worker, numworkers),
	}
}

func (w *WorkerPool) Start() {
	for i := 0; i < len(w.Workers); i++ {
		w.Workers[i] = NewWorker(w)
		w.Workers[i].Start()
	}
}

func (w *WorkerPool) Stop() {
	for i := 0; i < len(w.Workers); i++ {
		if w.Workers[i] != nil {
			w.Workers[i].Stop()
		}
	}
}
