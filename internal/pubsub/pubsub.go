package pubsub

type PubSub struct {
	TaskPubSub TaskPubSubInterface
}

func NewPubSub() *PubSub {
	return &PubSub{
		TaskPubSub: NewTaskPubSub(),
	}
}
