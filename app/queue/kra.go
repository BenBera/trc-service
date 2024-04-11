package queue

import (
	"bitbucket.org/maybets/kra-service/app/library"
	"bitbucket.org/maybets/kra-service/app/models"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (q *Queue) ProcessStake(deliveries <-chan amqp.Delivery) error {

	for d := range deliveries {
		data := models.KRAStakeInfo{}

		err := json.Unmarshal(d.Body, &data)
		log.Printf("%s", string(d.Body))

		if err != nil {

			log.Printf(" got error 1 decoding queue response to models.BetSettlement %s", err.Error())
			d.Ack(false)
			continue
		}
		log.Printf("Processing KRA STAKE Bets %s ", string(d.Body))
		err = q.submitStake(data)
		if err != nil {
			log.Printf(" got error decoding queue response to struct %s", err.Error())
			//d.Nack(false, false)
			d.Ack(false)
		}
		d.Ack(false)
	}

	return fmt.Errorf("handle: deliveries channel closed")
}

func (q *Queue) submitStake(data models.KRAStakeInfo) error {

	return library.SubmitStakeToKRA(q.DB, q.RedisConn, data)

}

func (q *Queue) ProcessOutcome(deliveries <-chan amqp.Delivery) error {

	for d := range deliveries {
		data := models.OutcomeInfo{}

		err := json.Unmarshal(d.Body, &data)
		log.Printf("%s", string(d.Body))

		if err != nil {

			log.Printf(" got error 1 decoding queue response to models.BetSettlement %s", err.Error())
			d.Ack(false)
			continue
		}
		log.Printf("Processing KRA STAKE Bets %s ", string(d.Body))
		err = q.submitOutcome(data)
		if err != nil {
			log.Printf(" got error decoding queue response to struct %s", err.Error())
			//d.Nack(false, false)
			d.Ack(false)
		}
		d.Ack(false)
	}

	return fmt.Errorf("handle: deliveries channel closed")
}

func (q *Queue) submitOutcome(data models.OutcomeInfo) error {

	return library.SubmitOutcomeToKRA(q.DB, q.RedisConn, data)
}

func (q *Queue) ProcessPRN(deliverie <-chan amqp.Delivery) error {
	return fmt.Errorf("handle: deliveries channel  ProcessTax closed")
}
func (q *Queue) processPRN(data models.GeneratePRN) error {
	dt, err := library.GeneratePaymentReferenceNumber(q.DB, q.RedisConn, data)
	if err != nil {
		log.Printf(fmt.Sprintf("Failed to generate prn %s", err.Error()))
		return err
	}
	log.Printf(fmt.Sprintf("Here are the prnRN Details %v", dt))
	return nil

}
