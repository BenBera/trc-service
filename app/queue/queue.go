package queue

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	trace "go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Queue struct {
	ConsumerConnection   *amqp.Connection
	PublisherQConnection *amqp.Connection
	DB                   *sql.DB
	RedisConn            *redis.Client
	//WalletServiceClient wallet.WalletClient
	//IdentityServiceClient identity.IdentityClient
	Tracer trace.Tracer
}

func (q *Queue) InitQueues() {

	ctx, span := q.Tracer.Start(context.Background(), "InitQueues")
	defer span.End()

	// get websocket instance
	queues := os.Getenv("queues")

	// get all the queues to consume from
	parts := strings.Split(queues, ",")
	sort.Strings(parts)

	// loop through the array
	for _, que := range parts {

		// remove any spaces
		que = strings.ToLower(strings.TrimSpace(que))

		if que == "kra_stake" {

			continue
		}

		// get number of workers/threads/consumers
		// convert into int
		numberOfWorkers, err := strconv.Atoi(os.Getenv(fmt.Sprintf("%s_workers", que)))
		if err != nil {

			numberOfWorkers = 1
			log.Printf(" cant retrieve %s workers got error %s ", que, err.Error())
		}

		x := 0

		for x < numberOfWorkers {

			// for each worker create a goroutine (thread/ run in background)
			go q.SetupQueue(ctx, que, 1)
			x++
		}

	}

	numberOfBetSettlementWorkers, _ := strconv.Atoi(os.Getenv("kra_stake_count"))
	betSettlementQueueOffset, _ := strconv.Atoi(os.Getenv("kra_stake_offset"))

	if betSettlementQueueOffset == 0 {

		go q.SetupQueue(ctx, "kra_stake", 100)

	}

	qNum := 0

	for qNum < numberOfBetSettlementWorkers {

		qu := fmt.Sprintf("bet_settlement.%d", qNum+betSettlementQueueOffset)
		go q.SetupQueue(ctx, qu, 100)
		qNum++
	}

	qNum = 0

	select {}

}
