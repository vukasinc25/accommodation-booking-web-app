package utility

import "github.com/nats-io/nats.go"

type Publisher interface {
	Publish(message interface{}) (*nats.Msg, error)
}

type Subscriber interface {
	Subscribe(function interface{}) error
}
