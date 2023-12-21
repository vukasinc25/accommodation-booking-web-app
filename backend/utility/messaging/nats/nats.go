package nats

import (
	"github.com/nats-io/nats.go"
	utility "github.com/vukasinc25/fst-airbnb/utility/messaging"
	"time"
)

type Publisher struct {
	conn    *nats.EncodedConn
	subject string
}

type Subscriber struct {
	conn    *nats.EncodedConn
	subject string
}

type AuthMessage struct {
	JToken string
}

func Connect() (*nats.Conn, error) {
	conn, err := nats.Connect("nats://pera:peric@nats:4222")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewNATSPublisher(subject string) (utility.Publisher, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	encConn, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}
	return &Publisher{
		conn:    encConn,
		subject: subject,
	}, nil
}

func NewNATSSubscriber(subject string) (utility.Subscriber, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}
	encConn, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}
	return &Subscriber{
		conn:    encConn,
		subject: subject,
	}, nil
}

func (p Publisher) Publish(message interface{}) (response *nats.Msg, err error) {
	err = p.conn.Request(p.subject, message, &response, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (s Subscriber) Subscribe(function interface{}) error {
	_, err := s.conn.Subscribe(s.subject, function)
	if err != nil {
		return err
	}
	return nil
}
