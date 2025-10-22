package main

import (
	"fmt"
	"strconv"
	"time"
	"wikicrawler/internal/utils/workers"
	tasks "wikicrawler/internal/utils/workers/task"
)

// Example target function
func ShortestPath(a int) {
	fmt.Println("Hello " + strconv.Itoa(a))
	time.Sleep(1000 * time.Millisecond)
}

// ---------- MAIN ----------
func main() {
	// Create worker pool with 3 workers and queue capacity 10
	workerPool := workers.NewWorkerPool(10, 100)

	// Start the worker pool
	workerPool.Start()

	// Start a goroutine to push tasks
	go func() {
		for i := 0; i < 20; i++ {
			task := tasks.NewTask(ShortestPath, i) // e.g. Add(i, i*2)
			ok := workerPool.Tasks.Push(task)
			if !ok {
				fmt.Printf("Task %d dropped: queue full\n", i)
			} else {
				fmt.Printf("Task %d pushed\n", i)
			}
		}
	}()

	// Let system run for a while
	time.Sleep(5 * time.Second)

	// Stop worker pool
	workerPool.Stop()
	fmt.Println("Worker pool stopped")
}
