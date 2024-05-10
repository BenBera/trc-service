package library

import (
	"bitbucket.org/maybets/kra-service/app/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	goutils "github.com/mudphilo/go-utils"
	"log"
	"os"
	"strconv"
)

// SubmitOutcomeToKRA sends outcome data to Kra  using models.OutcomeInfo as body request
func SubmitOutcomeToKRA(db *sql.DB, redisConn *redis.Client, data models.OutcomeInfo) error {

	login := GetSettings(db, redisConn, "KRA_LOGIN", "")
	transactionDate := TransactionDate()
	status := data.Status
	outcome := data.Outcome
	dateOfOutcome := data.Outcomedate
	// payout := data.Payout
	winnings := data.Winnings
	// winning_tax := data.WithholdingTax

	//get the bet details from bet ID
	winning_tax, err := strconv.ParseInt(data.Winnings, 10, 64)
	payout, err := strconv.ParseInt(data.Payout, 10, 64)




	var bets []models.OutcomeDetail
	if status == 1 {
		outcome = "LOSE"
	} else if status == 2 {
		outcome = "WON"
	} else if status == 5 {
		outcome = "CASHOUT"
	} else {
		outcome = ""
	}

	//Strings conversion to floats

	// Set them to the model
	outcomeInfo := models.OutcomeInfo{
		BetID:          data.BetID,
		Outcome:        outcome,
		Outcomedate:    dateOfOutcome,
		Payout:         fmt.Sprintf("%f", payout),
		Winnings:       fmt.Sprintf("%f", winnings),
		WithholdingTax: fmt.Sprintf("%f", winning_tax),
	}
	outcomeInfoWrapper := models.OutcomeDetail{
		OutcomeInfo: outcomeInfo,
	}

	bets = append(bets, outcomeInfoWrapper)

	noOfOutcomes := len(bets)

	//(pinNo+transactionDate +noOfStakes)
	hash := GetCheckSum(fmt.Sprintf("%s%s%d", login, transactionDate, noOfOutcomes))

	kraHeaders := models.OutcomeHeader{
		OperatorPin:     login,
		TransactionDate: transactionDate,
		NoOfOutcomes:    strconv.Itoa(noOfOutcomes),
	}
	requestPayload := models.OutcomeData{Request: models.OutcomeRequest{
		Hash:    hash,
		Header:  kraHeaders,
		Details: bets,
	}}

	KraEnvironment := GetSettings(db, redisConn, "KRA_ENVIRONMENT", "0")

	var endpoint string

	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)
	if KraEnvironmentInt == 1 {

		endpoint = os.Getenv("kra_production_submit_outcome_url")

	} else {

		endpoint = os.Getenv("kra_test_submit_outcome_url")

	}

	_, token := BearerToken(db, redisConn)

	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	var response models.KRAStakeResponse

	st, body := HTTPPost1(endpoint, header, requestPayload)

	if st > 300 || st < 200 {

		return fmt.Errorf(body)

	}

	_ = json.Unmarshal([]byte(body), &response)

	dbUtils := goutils.Db{DB: db}

	inserts := map[string]interface{}{
		"bet_id":             data.BetID,
		"status":             response.Response.Result.ResponseCode,
		"status_description": response.Response.Result.Message,
		"created":            goutils.MysqlNow(),
	}

	_, err := dbUtils.Upsert("outcome_remittance", inserts, nil)
	if err != nil {

		log.Printf("error creating outcome_remittance %s ", err.Error())
		return err
	}

	return nil

}

// SubmitStakeToKRA Sends Stake date to kra using models.KRAStakeInfo as body request
func SubmitStakeToKRA(db *sql.DB, redisConn *redis.Client, data models.KRAStakeInfo) error {

	login := GetSettings(db, redisConn, "KRA_LOGIN", "")
	transactionDate := TransactionDate()
	log.Printf(fmt.Sprintf("Here is the bet_id SubmitBet file %d", data.BetID))
	var bets []models.StakeDetail

	// Set them to the model
	stakeinfo := models.KRAStakeInfo{
		BetID:       data.BetID,
		CustomerID:  data.CustomerID,
		MobileNo:    data.MobileNo,
		StakeAmt:    data.StakeAmt,
		PunterAmt:   data.PunterAmt,
		Odds:        data.Odds,
		ExciseAmt:   data.ExciseAmt,
		Desc:        data.Desc,
		StakeType:   data.StakeType,
		DateOfStake: data.DateOfStake,
	}
	stakeinfoWrapper := models.StakeDetail{
		StakeInfo: stakeinfo,
	}

	bets = append(bets, stakeinfoWrapper)

	log.Printf(fmt.Sprintf("Here is the stake information %s", bets))
	noOfStakes := len(bets)
	log.Printf(fmt.Sprintf("Here is the number of Stakes %d", noOfStakes))
	//(pinNo+transactionDate +noOfStakes)
	hash := GetCheckSum(fmt.Sprintf("%s%s%d", login, transactionDate, noOfStakes))
	//hash = fmt.Sprintf("%sinvalid",hash)

	kraHeaders := models.KRAHeader{
		OperatorPin:     login,
		TransactionDate: transactionDate,
		NoOfStakes:      strconv.Itoa(noOfStakes),
	}

	requestPayload := models.StakeData{Request: models.KRARequest{
		Hash:    hash,
		Header:  kraHeaders,
		Details: bets,
	}}

	KraEnvironment := GetSettings(db, redisConn, "KRA_ENVIRONMENT", "0")

	var endpoint string

	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)
	if KraEnvironmentInt == 1 {

		endpoint = os.Getenv("kra_production_submit_stake_url")

	} else {

		endpoint = os.Getenv("kra_test_submit_stake_url")

	}

	_, token := BearerToken(db, redisConn)
	//token = fmt.Sprintf("%sinvalid",token)

	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	var response models.KRAStakeResponse

	st, body := HTTPPost(endpoint, header, requestPayload)

	if st > 300 || st < 200 {

		return fmt.Errorf(body)

	}

	_ = json.Unmarshal([]byte(body), &response)

	dbUtils := goutils.Db{DB: db}

	inserts := map[string]interface{}{
		"bet_id":             data.BetID,
		"status":             response.Response.Result.ResponseCode,
		"status_description": response.Response.Result.Message,
		"created":            goutils.MysqlNow(),
	}
	_, err := dbUtils.Upsert("stake_remittance", inserts, nil)
	if err != nil {

		log.Printf("error creating stake_remittance %s ", err.Error())
		return err
	}

	return nil

}

//PayExcise

func PayExcise(db *sql.DB, data models.PayTax) {

	//prn
	//amount

}

//PayWithHolding

func PayWithHolding() {

}
