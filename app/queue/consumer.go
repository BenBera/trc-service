package queue

import (
	"bitbucket.org/maybets/kra-service/app/constants"
	"bitbucket.org/maybets/kra-service/app/rabbitmq"
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"os"
	"strings"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

func (q *Queue) RouteMessage(ctx context.Context, deliveries <-chan amqp.Delivery, queue string) error {

	ctx, span := q.Tracer.Start(ctx, "RouteMessage", oteltrace.WithAttributes(attribute.String("queueName", queue)))
	defer span.End()

	//log.Printf("route message for queue %s", queue)

	prefix := os.Getenv("queue_prefix")

	if len(prefix) > 0 {

		parts := strings.Split(queue, ".")
		if len(parts) > 1 {

			queue = strings.Join(parts[1:], ".")
		}
	}

	queue = strings.ToLower(queue)

	if queue == "kra_stake" {

		return q.ProcessStake(deliveries)

	} else if strings.HasPrefix(queue, "kra_stake") {

		return q.ProcessStake(deliveries)

	} else if queue == "kra_outcome" {

		return q.ProcessOutcome(deliveries)

	} else if strings.HasPrefix(queue, "kra_outcome") {
		return q.ProcessOutcome(deliveries)
	}
	return nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {

		return fmt.Errorf("consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {

		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	// wait for handle() to exit
	return <-c.done
}

func (q *Queue) SetupQueue(ctx context.Context, QueueName string, prefetchCount int) {

	forever := make(chan bool)

	conn := rabbitmq.NewConnection(q.Tracer, ctx, QueueName, 5, prefetchCount)

	if err := conn.Connect(ctx, q.ConsumerConnection); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error during conn.Connect ",
				constants.DATA:        QueueName,
			}).
			Panic(err.Error())

	}

	if err := conn.BindQueue(ctx); err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error during conn.BindQueue  ",
				constants.DATA:        QueueName,
			}).
			Panic(err.Error())
	}

	err := conn.Consume(ctx, q.RouteMessage)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error during conn.Consume  ",
				constants.DATA:        QueueName,
			}).
			Panic(err.Error())
	}

	<-forever

}
