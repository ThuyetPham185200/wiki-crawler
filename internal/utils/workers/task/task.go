package tasks

import (
	"reflect"
	"time"
)

type Task struct {
	Func    interface{}   // function to call
	Args    []interface{} // arguments to pass
	Results []interface{} // Result to return
}

func NewTask(f interface{}, args ...interface{}) *Task {
	return &Task{
		Func: f,
		Args: args,
	}
}

// DoTask executes the stored task
func (t *Task) DoTask() {
	values := make([]reflect.Value, len(t.Args))
	for i := range t.Args {
		values[i] = reflect.ValueOf(t.Args[i])
	}
	results := reflect.ValueOf(t.Func).Call(values)

	t.Results = make([]interface{}, len(results))
	for i, r := range results {
		t.Results[i] = r.Interface()
	}
}

type TaskQueue struct {
	Capacity int
	Tasks    chan *Task
}

func NewTaskQueue(cap int) *TaskQueue {
	return &TaskQueue{
		Capacity: cap,
		Tasks:    make(chan *Task, cap),
	}
}

func (q *TaskQueue) Push(t *Task) bool {
	select {
	case q.Tasks <- t:
		return true
	default:
		return false
	}
}

func (q *TaskQueue) TryPop(timeout time.Duration) *Task {
	select {
	case t := <-q.Tasks:
		return t
	case <-time.After(timeout):
		return nil
	}
}

func (q *TaskQueue) Close() {
	close(q.Tasks)
}
