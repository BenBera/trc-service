package rabbitmq

import (
	"bitbucket.org/maybets/kra-service/app/constants"
	"context"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"os"
	"strings"
)

// MessageBody is the struct for the body passed in the AMQP message. The type will be set on the Request header
type MessageBody struct {
	Data []byte
	Type string
}

// Message is the amqp request to publish
type Message struct {
	Queue         string
	ReplyTo       string
	ContentType   string
	CorrelationID string
	Priority      uint8
	Body          []byte
}

// RabbitMQConnection is the connection created
type RabbitMQConnection struct {
	name          string
	conn          *amqp.Connection
	channel       *amqp.Channel
	exchange      string
	queue         string
	routingKey    string
	PrefetchCount int
	MaxPriority   int
	ConsumerTag   string
	err           chan error
	Tracer        trace.Tracer
}

var (
	connectionPool = make(map[string]*RabbitMQConnection)
)

// NewConnection returns the new connection object
func NewConnection(r trace.Tracer, ctx context.Context, queueName string, maxPriority, PrefetchCount int) *RabbitMQConnection {

	ctx, span := r.Start(ctx, "NewConnection", oteltrace.WithAttributes(attribute.String("queueName", queueName)))
	defer span.End()

	prefix := os.Getenv("queue_prefix")
	if len(prefix) > 0 {

		queueName = fmt.Sprintf("%s.%s", prefix, queueName)
	}

	queueName = strings.ToLower(queueName)

	ConsumerTag := queueName
	queue := queueName
	exchange := queueName
	routingKey := queueName

	c := &RabbitMQConnection{
		exchange:      exchange,
		queue:         queue,
		routingKey:    routingKey,
		PrefetchCount: PrefetchCount,
		err:           make(chan error),
		MaxPriority:   maxPriority,
		ConsumerTag:   ConsumerTag,
		Tracer:        r,
	}

	return c

}

// GetConnection returns the connection which was instantiated
func GetConnection(name string) *RabbitMQConnection {

	return connectionPool[name]

}

func (r *RabbitMQConnection) Connect(ctx context.Context, conn *amqp.Connection) error {

	ctx, span := r.Tracer.Start(ctx, "Connect")
	defer span.End()

	var err error

	r.conn = conn

	r.channel, err = r.conn.Channel()
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error opening channel conn.Channel ",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return fmt.Errorf("channel: %s", err.Error())

	}

	go func() {

		<-r.channel.NotifyClose(make(chan *amqp.Error)) //Listen to NotifyClose

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: fmt.Sprintf("%s channel is closed", r.name),
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		r.err <- errors.New(fmt.Sprintf("%s channel closed", r.name))

	}()

	// set channel properties
	err = r.channel.Qos(
		r.PrefetchCount, // prefetch count
		0,               // prefetch size
		false,           // global
	)

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "exchange Qos error",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return fmt.Errorf("channel: %s", err)
	}

	if err := r.channel.ExchangeDeclare(
		r.exchange, // name
		"direct",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // noWait
		nil,        // arguments
	); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "exchange declare error",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return fmt.Errorf("error in Exchange Declare: %s", err.Error())
	}

	return nil
}

func (r *RabbitMQConnection) BindQueue(ctx context.Context) error {

	ctx, span := r.Tracer.Start(ctx, "BindQueue")
	defer span.End()

	var args amqp.Table
	args = amqp.Table{"x-max-priority": int32(r.MaxPriority)}

	if _, err := r.channel.QueueDeclare(r.queue, true, false, false, false, args); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error in declaring the queue channel.QueueDeclare ",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return fmt.Errorf("error in declaring the queue %s", err)

	}

	if err := r.channel.QueueBind(r.queue, r.routingKey, r.exchange, false, nil); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error in declaring the queue channel.QueueBind ",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return fmt.Errorf("Queue  Bind error: %s", err)

	}

	return nil
}

// Reconnect reconnects the connection
func (r *RabbitMQConnection) Reconnect(ctx context.Context) error {

	ctx, span := r.Tracer.Start(ctx, "Reconnect")
	defer span.End()

	logrus.WithContext(ctx).
		WithFields(logrus.Fields{
			constants.DESCRIPTION: "reconnecting ",
			constants.DATA:        r.queue,
		}).
		Infof("reconnection %s channel ", r.name)

	if err := r.Connect(ctx, r.conn); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "failed to Reconnect consumer  ",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return err
	}

	if err := r.BindQueue(ctx); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "failed to BindQueue  ",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return err
	}

	return nil

}

// Consume consumes the messages from the queues and passes it as map of chan of amqp.Delivery
func (r *RabbitMQConnection) Consume(ctx context.Context, fn func(context.Context, <-chan amqp.Delivery, string) error) error {

	ctx, span := r.Tracer.Start(ctx, "Consume")
	defer span.End()

	delivery, err := r.channel.Consume(r.queue, r.queue, false, false, false, false, nil)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error starting to consume  ",
				constants.DATA:        r.queue,
			}).
			Error(err.Error())

		return err
	}

	for {

		err = fn(ctx, delivery, r.ConsumerTag)

		if err := <-r.err; err != nil {

			r.Reconnect(ctx)

			err := r.Consume(ctx, fn)
			if err != nil {

				logrus.WithContext(ctx).
					WithFields(logrus.Fields{
						constants.DESCRIPTION: "failed to setup consumer  ",
						constants.DATA:        r.queue,
					}).
					Error(err.Error())

			}
		}

	}

	//return nil

}
