package pubsub

import (
	_model "bitbucket.org/edts/go-task-management/internal/model"
	_logger "bitbucket.org/edts/go-task-management/pkg/logger"
	"sync"
)

var logs = _logger.GetContextLoggerf(nil)

type TaskPubSubInterface interface {
	Subscribe(teamID string) <-chan TaskEvent
	Publish(teamID, eventType string, task *_model.Task)
	Unsubscribe(teamID string, ch <-chan TaskEvent)
}

// TaskEvent define task event types
type TaskEvent struct {
	Type string // "created" or "updated"
	Task *_model.Task
}

// TaskPubSub manages task related events
type TaskPubSub struct {
	mu          sync.Mutex
	subscribers map[string][]chan TaskEvent
}

// NewTaskPubSub init TaskPubSub
func NewTaskPubSub() *TaskPubSub {
	return &TaskPubSub{
		subscribers: make(map[string][]chan TaskEvent),
	}
}

// Subscribe to task events
func (ps *TaskPubSub) Subscribe(teamID string) <-chan TaskEvent {
	logs.Infof("Subscribe:: Start subscribing task of teamId: %s", teamID)
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan TaskEvent, 1) // Set buffered channel
	ps.subscribers[teamID] = append(ps.subscribers[teamID], ch)
	logs.Info("Subscribe:: Finish subscribing the task")
	return ch
}

// Publish a task event and notify all subs
func (ps *TaskPubSub) Publish(teamID, eventType string, task *_model.Task) {
	logs.Infof("Publish:: Start publishing task with type:%s - %v", eventType, task)
	ps.mu.Lock()
	defer ps.mu.Unlock()

	event := TaskEvent{Type: eventType, Task: task}

	// Notify all subscribers for the team
	for _, ch := range ps.subscribers[teamID] {
		ch <- event
	}
	logs.Info("Publish:: Finish notifying the published task")
}

// Unsubscribe from task events
func (ps *TaskPubSub) Unsubscribe(teamID string, ch <-chan TaskEvent) {
	logs.Infof("Unsubscribe:: Start unsubscribe task of teamId: %s", teamID)
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for i, sub := range ps.subscribers[teamID] {
		// Find the channel that the subscriber wants to remove
		if sub == ch {
			// Remove the channel
			ps.subscribers[teamID] = append(ps.subscribers[teamID][:i], ps.subscribers[teamID][i+1:]...)
			close(sub)
			break
		}
	}
	logs.Info("Unsubscribe:: Finish unsubscribe (ack) the task")
}
