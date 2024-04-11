package controllers

import (
	"bitbucket.org/maybets/kra-service/app/library"
	"bitbucket.org/maybets/kra-service/app/models"
	"database/sql"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/logrusorgru/aurora"
	goutils "github.com/mudphilo/go-utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"strings"
)

type Controller struct {
	//Conn rabbitMQ connection
	Conn *amqp.Connection
	//ArriveTime when the request arrived, use to measure api response time
	ArriveTime int64
	//RedisConnection redis connection
	RedisConnection *redis.Client
	APICache        *redis.Client
	//DB database connection
	DB *sql.DB
	//E echo server
	E *echo.Echo
	// MQTT Broker
	MQTT *mqtt.Client
	//Encryption Keys
	Key    *string
	IV     *string
	Tracer trace.Tracer
	//IdentityServiceClient identity.IdentityClient
	//BettingServiceClient betting.BettingClient
	//JackpotServiceClient  jackpot.JackpotClient
	//BonusServiceClient    bonus.BonusClient
}

func GetSettings(db *sql.DB) models.Settings {

	dbUtil := goutils.Db{DB: db}

	sqlQuery := "SELECT withdraw_status," +
		" withdrawal_minimum_amount," +
		" withdrawal_maximum_amount," +
		" withdrawal_daily_limit," +
		" withdrawal_before_minimum_spend, " +
		" cannot_withdraw_more_than_cumulative_deposit," +
		" cannot_withdraw_more_than_cumulative_stake," +
		" cannot_withdraw_more_than_cumulative_winning" +
		" FROM settings WHERE id = 1 "

	dbUtil.SetQuery(sqlQuery)

	var withdrawStatus,
		withdrawalMinimumAmount,
		withdrawalMaximumAmount,
		withdrawalDailyLimit,
		withdrawalBeforeMinimumSpend,
		cannotWithdrawMoreThanCumulativeDeposit,
		cannotWithdrawMoreThanCumulativeStake,
		cannotWithdrawMoreThanCumulativeWinning sql.NullInt64

	err := dbUtil.FetchOne().Scan(&withdrawStatus,
		&withdrawalMinimumAmount,
		&withdrawalMaximumAmount,
		&withdrawalDailyLimit,
		&withdrawalBeforeMinimumSpend,
		&cannotWithdrawMoreThanCumulativeDeposit,
		&cannotWithdrawMoreThanCumulativeStake,
		&cannotWithdrawMoreThanCumulativeWinning)
	if err != nil {

		log.Printf("error retrieving bonus | %s ", err.Error())

		return models.Settings{}
	}

	re := models.Settings{
		WithdrawStatus:                          withdrawStatus.Int64,
		WithdrawalMinimumAmount:                 withdrawalMinimumAmount.Int64,
		WithdrawalMaximumAmount:                 withdrawalMaximumAmount.Int64,
		WithdrawalDailyLimit:                    withdrawalDailyLimit.Int64,
		WithdrawalBeforeMinimumSpend:            withdrawalBeforeMinimumSpend.Int64,
		CannotWithdrawMoreThanCumulativeDeposit: cannotWithdrawMoreThanCumulativeDeposit.Int64,
		CannotWithdrawMoreThanCumulativeStake:   cannotWithdrawMoreThanCumulativeStake.Int64,
		CannotWithdrawMoreThanCumulativeWinning: cannotWithdrawMoreThanCumulativeWinning.Int64,
	}

	return re

}

func RespondJSON(c echo.Context, code int, message interface{}) error {

	return c.JSON(code, models.ResponseMessage{
		Status:  code,
		Message: message,
	})
}

// respondError makes the error response with payload as json format
func RespondRaw(c echo.Context, code int, message interface{}) error {
	return c.JSON(code, message)
}
func GetJSONRawBody(c echo.Context) map[string]interface{} {

	request := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {

		//log.Printf("empty json body %s ",err.Error())
		return nil
	}

	return request
}

func GetRawBody(c echo.Context) string {

	buf := new(strings.Builder)
	_, err := io.Copy(buf, c.Request().Body)
	if err != nil {

		//log.Error("empty json body")
		return ""
	}

	return buf.String()
}
func GetSMSTemplate(db *sql.DB, conn *redis.Client, templateName string) string {

	// retrieve SMS template
	templateName = fmt.Sprintf("SMS:TEMPLATE:%s", templateName)

	sms, err := library.GetRedisKey(conn, templateName)

	if err != nil || len(sms) == 0 {

		//log.Printf(" got error retrieving SMS template %s error %s",templateName,err.Error())

		var tName sql.NullString
		err = db.QueryRow("SELECT message FROM sms_template WHERE name = ? LIMIT 1 ", templateName).Scan(&tName)
		if err != nil {

			log.Printf("%s", aurora.Red(fmt.Sprintf(" got error retriving sms template for %s error %s", templateName, err.Error())))
			sms = ""

		} else if !tName.Valid {

			log.Printf("%s", aurora.Red(fmt.Sprintf(" got error retriving sms template for %s template does not exist", templateName)))
			sms = ""

		} else {

			sms = tName.String

		}

		err = library.SetRedisKey(conn, templateName, sms)
		if err != nil {

			log.Printf("got error saving SMS Templates")
		}
	}

	if strings.ToLower(sms) == "none" || strings.ToLower(sms) == "na" || strings.ToLower(sms) == "n/a" {

		return ""
	}

	return sms

}
func IsIPBlocked(db *sql.DB, ip_address string) bool {
	return false

	var check sql.NullInt64

	err := db.QueryRow("SELECT COUNT(id) as checks FROM ip_address_blacklist WHERE ip_address = ? ", ip_address).Scan(&check)
	if err == sql.ErrNoRows {

		return false
	}

	if err != nil {

		log.Printf("Got error checking if IP is blocked %s ", err.Error())
		return true
	}

	return !check.Valid || check.Int64 > 0
}
