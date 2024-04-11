package crontask

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	_ "github.com/sirupsen/logrus"
	trace "go.opentelemetry.io/otel/trace"
)

type Crontask struct {
	RabbitMQConn *amqp.Connection
	DB           *sql.DB
	RedisConn    *redis.Client
	//WalletServiceClient wallet.WalletClient
	Tracer trace.Tracer
}

func (cron *Crontask) SetupJobs() {

	// Start a new root span representing the entire request.
	ctx, span := cron.Tracer.Start(context.Background(), "kra_cron-task")
	defer span.End()

	go cron.SendDashboardReports(ctx)
	select {}
}

func PayTax(cro *Crontask) {
	//get data from dab
	//confirm the date
	// pass the b2b

}
