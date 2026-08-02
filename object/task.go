package object

import "sync"

// Task is an opaque handle for one concurrently evaluated Silver expression.
// Its result is published exactly once and may be collected exactly once.
type Task struct {
	done chan struct{}

	mu        sync.Mutex
	result    Object
	collected bool
	name      string
}

// NewTask constructs a pending task handle.
func NewTask() *Task { return &Task{done: make(chan struct{})} }

func (t *Task) Type() ObjectType { return TASK_OBJ }
func (t *Task) Inspect() string  { return "<task>" }

// Complete publishes the task result and wakes every waiter.
func (t *Task) Complete(result Object) {
	t.mu.Lock()
	t.result = result
	close(t.done)
	t.mu.Unlock()
}

// MarkCollected atomically consumes the handle without waiting for its work.
func (t *Task) MarkCollected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.collected {
		return false
	}
	t.collected = true
	return true
}

// Await blocks until completion and returns the published value.
func (t *Task) Await() Object {
	<-t.done
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.result
}

func (t *Task) Collected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.collected
}

func (t *Task) SetName(name string) {
	t.mu.Lock()
	if t.name == "" {
		t.name = name
	}
	t.mu.Unlock()
}

func (t *Task) Name() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.name
}
