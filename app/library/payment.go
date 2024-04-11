package library

import (
	"bitbucket.org/maybets/kra-service/app/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/labstack/gommon/log"
	goutils "github.com/mudphilo/go-utils"
	"os"
	"strconv"
	"strings"
)

type KRAPRN struct {
	Response struct {
		Result struct {
			ResponseCode string `json:"ResponseCode"`
			Message      string `json:"Message"`
			Status       string `json:"Status"`
			PrnDetails   struct {
				Prn        string `json:"PRN"`
				PrnRegDate string `json:"prnRegDate"`
				PrnAmount  string `json:"prnAmount"`
				PrnExpDate string `json:"prnExpDate"`
			} `json:"prnDetails"`
		} `json:"RESULT"`
	} `json:"RESPONSE"`
}

type MPESAB2BPaymentRequest struct {
	InitiatorName          string `json:"Initiator"`
	SecurityCredential     string `json:"SecurityCredential"`
	CommandID              string `json:"CommandID"`
	Amount                 int64  `json:"Amount"`
	PartyA                 int64  `json:"PartyA"`
	PartyB                 int64  `json:"PartyB"`
	SenderIdentifier       string `json:"SenderIdentifier"`
	RecieverIdentifierType string `json:"RecieverIdentifierType"`
	Remarks                string `json:"Remarks"`
	QueueTimeOutURL        string `json:"QueueTimeOutURL"`
	ResultURL              string `json:"ResultURL"`
	Occassion              string `json:"Occassion"`
	AccountReference       string `json:"AccountReference"`
}

type MPESAB2CPaymentRequest struct {
	InitiatorName          string `json:"Initiator"`
	SecurityCredential     string `json:"SecurityCredential"`
	CommandID              string `json:"CommandID"`
	SenderIdentifierType   string `json:"SenderIdentifierType"`
	RecieverIdentifierType string `json:"RecieverIdentifierType"`
	Amount                 string `json:"Amount"`
	PartyA                 string `json:"PartyA"`
	PartyB                 string `json:"PartyB"`
	AccountReference       string `json:"AccountReference"`
	Remarks                string `json:"Remarks"`
	QueueTimeOutURL        string `json:"QueueTimeOutURL"`
	ResultURL              string `json:"ResultURL"`
}

type MPESAB2BPaymentRequestResponse struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ConversationID           string `json:"ConversationID"`
	ResponseCode             string `json:"ResponseCode"`
	ResponseDescription      string `json:"ResponseDescription"`
}

// MpesaAccessTokenResponseB2C is the response sent back by Safaricom when we make a request to generate a token
type MpesaAccessTokenResponseB2C struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RequestID    string `json:"requestId"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// generateAccessToken sends a http request to generate new access token
func GenerateB2CAccessToken(db *sql.DB, conn *redis.Client) (*string, error) {

	var endpoint string
	var redisKey string

	KraEnvironment := GetSettings(db, conn, "KRA_ENVIRONMENT", "0")
	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)
	if KraEnvironmentInt == 1 {

		endpoint = os.Getenv("kra_production_b2b_access_token_url")
		redisKey = "B2C:ACCESS_TOKEN"

	} else {

		endpoint = os.Getenv("kra_test_b2b_access_token_url")
		redisKey = "B2C:ACCESS_TOKEN:TEST"
	}

	token, err := GetRedisKey(conn, redisKey)
	if err != nil {

		log.Printf("error getting B2C:ACCESS_TOKEN from redis %s ", err.Error())

	} else {

		if len(token) > 2 {

			return &token, nil
		}
	}

	consumerKey := GetSettings(db, conn, "B2C_CONSUMER_KEY", "")
	secretKey := GetSettings(db, conn, "B2C_CONSUMER_SECRET", "")
	log.Printf(fmt.Sprintf("Here is the consumerKey %s", consumerKey))
	log.Printf(fmt.Sprintf("Here is the consumerSecret %s", secretKey))
	log.Printf(fmt.Sprintf("KRA Environment %s", KraEnvironment))
	log.Printf(fmt.Sprintf("KRA Endpoint %s", endpoint))

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "application/json",
		"Authorization": fmt.Sprintf("Basic " + ToBase64Token(fmt.Sprintf("%s:%s", consumerKey, secretKey))),
	}

	httpBody, httpStatus, _ := HTTPGetWithHeaders(endpoint, headers, nil)
	if httpStatus > 300 || httpStatus < 200 {

		return nil, fmt.Errorf(httpBody)
	}

	accessTokenResponse := new(MpesaAccessTokenResponseB2C)

	if err := json.Unmarshal([]byte(httpBody), &accessTokenResponse); err != nil {

		log.Printf("Error retrieving B2C Token %s", err.Error())
		return nil, err
	}

	token = accessTokenResponse.AccessToken
	log.Printf("got %s - %s", redisKey, token)

	err = SetRedisKeyWithExpiry(conn, redisKey, token, 15*60)
	if err != nil {

		log.Printf("Error saving B2C Token %s", err.Error())

	}

	return &token, nil
}

func GeneratePaymentReferenceNumber(db *sql.DB, redisConn *redis.Client, pin models.GeneratePRN) (models.PayTax, error) {
	data := models.PayTax{}
	KraRegistrationStatus := os.Getenv("KRA_REGISTRATION_STATUS")
	KraRegistrationStatusInt, _ := strconv.ParseInt(KraRegistrationStatus, 10, 64)
	if KraRegistrationStatusInt == 0 {

		return models.PayTax{}, fmt.Errorf("kra is not yet setup")
	}

	KraEnvironment := os.Getenv("KRA_ENVIRONMENT")

	var endpoint string

	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)
	if KraEnvironmentInt == 1 {

		log.Printf("Here is the Environment On productions", KraEnvironment)
		endpoint = os.Getenv("prod_url")

	} else {

		endpoint = os.Getenv("staging_url")

	}

	login := os.Getenv("KRA_PIN")
	log.Printf("Here is the Pin ", login)
	loginvV := strings.ToUpper(login)
	log.Printf("Here is the Pin ", loginvV)

	transactionDate := TransactionDate()

	hash := GetCheckSum(fmt.Sprintf("%s%s%s", login, transactionDate, pin.TaxType))
	//hash = fmt.Sprintf("%sinvalid",hash)

	payload := map[string]interface{}{
		"Request": map[string]interface{}{
			"hash": hash,
			"paymentInfo": map[string]interface{}{
				"pinNo":           loginvV,
				"amount":          pin.Amount,
				"taxType":         pin.TaxType,
				"transactionDate": transactionDate,
				"periodFrom":      pin.StartDate,
				"periodTo":        pin.EndDate,
			},
		},
	}

	_, token := BearerToken(db, redisConn)
	//token = fmt.Sprintf("%sinvalid",token)

	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	var response KRAPRN

	st, body := HTTPPost(endpoint, header, payload)

	if st > 300 || st < 200 {

		//library.Notification(db, cfg, redisConn, "KRA-STATUS", "Failed to generate KRA PRN", fmt.Sprintf("Got invalid status during KRA PRN Generation\npin %s\namount %0.2f\nTax Type %s\nstatusCode %d,response%s", login, amount, taxType, st, body))
		return models.PayTax{}, fmt.Errorf(body)

	}

	_ = json.Unmarshal([]byte(body), &response)

	if response.Response.Result.ResponseCode != "1111" {

		//library.Notification(db, cfg, redisConn, "KRA-STATUS", "Failed to generate KRA PRN", fmt.Sprintf("Got invalid status during KRA PRN Generation\npin %s\namount %0.2f\nTax Type %s\nstatusCode %d,response%s", login, amount, taxType, st, response.Response.Result.Message))
		return models.PayTax{}, fmt.Errorf(response.Response.Result.Message)
	}

	dbUtils := goutils.Db{DB: db}

	inserts := map[string]interface{}{
		"start_date":               pin.StartDate,
		"end_date":                 pin.EndDate,
		"tax_type":                 pin.TaxType,
		"transaction_date":         transactionDate,
		"amount":                   pin.Amount,
		"payment_reference_number": response.Response.Result.PrnDetails.Prn,
		"status":                   0,
		"status_description":       "Pending Payment",
		"mpesa_reference":          "",
		"created":                  goutils.MysqlNow(),
	}

	transactionID, err := dbUtils.Upsert("tax_remittance", inserts, nil)
	if err != nil {

		log.Printf("error creating tax_remittance %s ", err.Error())
		return models.PayTax{}, err
	}
	data = models.PayTax{
		Amount:        pin.Amount,
		Prn:           response.Response.Result.PrnDetails.Prn,
		TransactionID: transactionID,
	}
	err = SendB2B(db, redisConn, data)
	if err != nil {
		return models.PayTax{}, err
	}

	return data, nil

}

func SendB2B(db *sql.DB, redisConn *redis.Client, data models.PayTax) error {

	KraEnvironment := GetSettings(db, redisConn, "KRA_ENVIRONMENT", "0")

	var endpoint string
	var kraPayBill string
	var resultUrl string

	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)

	isProduction := false
	if KraEnvironmentInt == 1 {

		endpoint = os.Getenv("kra_b2b_url")
		kraPayBill = os.Getenv("kra_production_paybill")
		resultUrl = os.Getenv("kra_production_result_url")
		isProduction = true

	} else {

		endpoint = os.Getenv("kra_test_b2b_url")
		kraPayBill = os.Getenv("kra_test_paybill")
		resultUrl = os.Getenv("kra_production_result_url")
		isProduction = false
	}

	resultUrl = strings.ReplaceAll(resultUrl, "{transactionID}", fmt.Sprintf("%d", data.TransactionID))

	initiator := GetSettings(db, redisConn, "B2C_INITIATOR", "")
	password := GetSettings(db, redisConn, "B2C_PASSWORD", "")
	paybill := GetSettings(db, redisConn, "B2C_PAYBILL", "0")

	securityCredentials := GetSettings(db, redisConn, "B2C_ENCRYPTED_PASSWORD", "")

	if len(securityCredentials) == 0 {

		creds, err := GenerateSecurityCredentials(password, isProduction)
		if err != nil {

			log.Printf("Got Error generating securityCredentials  %s", err.Error())
			return err
		}

		securityCredentials = creds

	}

	payload := MPESAB2CPaymentRequest{
		InitiatorName:          initiator,
		SecurityCredential:     securityCredentials,
		CommandID:              "PayTaxToKRA",
		Amount:                 fmt.Sprintf("%d", int64(data.Amount)),
		PartyA:                 paybill,
		PartyB:                 kraPayBill,
		Remarks:                data.Prn,
		AccountReference:       data.Prn,
		QueueTimeOutURL:        resultUrl,
		ResultURL:              resultUrl,
		SenderIdentifierType:   "4",
		RecieverIdentifierType: "4",
	}

	accessToken, err := GenerateB2CAccessToken(db, redisConn)
	if err != nil {

		log.Printf(fmt.Sprintf("Got Error generating GenerateB2CAccessToken  %s", err.Error()))
		return err
	}

	authorization := fmt.Sprintf("Bearer %s", *accessToken)

	headers := map[string]string{
		"Authorization": authorization,
	}

	httpStatus, httpBody := HTTPPost(endpoint, headers, payload)
	log.Printf("Send B2B Tax status %d response %s ", httpStatus, httpBody)

	response := new(MPESAB2BPaymentRequestResponse)

	err = json.Unmarshal([]byte(httpBody), response)

	updates := map[string]interface{}{
		"mpesa_response_code":              response.ResponseCode,
		"mpesa_conversation_id":            response.ConversationID,
		"mpesa_originator_conversation_id": response.OriginatorConversationID,
		"mpesa_response_description":       response.ResponseDescription,
		"status":                           1,
		"status_description":               "Payment send to Safaricom",
	}

	andWhere := map[string]interface{}{
		"id": data.TransactionID,
	}

	dbUtils := goutils.Db{DB: db}

	_, err = dbUtils.Update("tax_remittance", andWhere, updates)
	if err != nil {

		log.Printf(fmt.Sprintf("failed to update tax_remittance %s ", err.Error()))
	}
	//message := fmt.Sprintf("Succesfully paid To KRA for prn %s and amount of %0.2f", prnNumber, amount)
	//library.SendSMS(db, cfg, redisConn, 254708491516, message, "ALERT")

	return nil
}
