package library

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"os"
)

// MpesaB2C is an application that will be making a transaction
type MpesaB2C struct {
	consumerKey    string
	consumerSecret string
	baseURL        string
	client         *http.Client
}
type TransR struct {
	baseURL string
	client  *http.Client
}

// MpesaOptsB2C stores all the configuration keys we need to set up a Mpesa app,
type MpesaOptsB2C struct {
	ConsumerKey    string
	ConsumerSecret string
	BaseURL        string
}

// B2CRequestBody is the body with the parameters to be used to initiate a B2C request
type B2CRequestBody struct {
	InitiatorName      string `json:"InitiatorName"`
	SecurityCredential string `json:"SecurityCredential"`
	CommandID          string `json:"CommandID"`
	Amount             string `json:"Amount"`
	PartyA             string `json:"PartyA"`
	PartyB             string `json:"PartyB"`
	Remarks            string `json:"Remarks"`
	QueueTimeOutURL    string `json:"QueueTimeOutURL"`
	ResultURL          string `json:"ResultURL"`
	Occassion          string `json:"Occassion"`
}
type TransactionalRequestBody struct {
	Initiator          string `json:"Initiator"`
	SecurityCredential string `json:"SecurityCredential"`
	CommandID          string `json:"CommandID"`
	TransactionID      string `json:"TransactionID"`
	PartyA             int    `json:"PartyA"`
	IdentifierType     int    `json:"IdentifierType"`
	ResultURL          string `json:"ResultURL"`
	QueueTimeOutURL    string `json:"QueueTimeOutURL"`
	Remarks            string `json:"Remarks"`
	Occassion          string `json:"Occassion"`
}

// B2CRequestResponse is the response sent back after initiating a B2C request.
type B2CRequestResponse struct {
	ConversationID           string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ResponseCode             string `json:"ResponseCode"`
	ResponseDescription      string `json:"ResponseDescription"`
	RequestID                string `json:"requestId"`
	ErrorCode                string `json:"errorCode"`
	ErrorMessage             string `json:"errorMessage"`
}
type TransactionStatusResponse struct {
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ConversationID           string `json:"ConversationID"`
	ResponseCode             string `json:"ResponseCode"`
	ResponseDescription      string `json:"ResponseDescription"`
	RequestID                string `json:"requestId"`
	ErrorCode                string `json:"errorCode"`
	ErrorMessage             string `json:"errorMessage"`
}

// makeRequest performs all the http requests for the specific app
func (m *MpesaB2C) makeRequest(req *http.Request) ([]byte, error) {

	resp, err := m.client.Do(req)
	if resp != nil {

		defer resp.Body.Close()
	}
	if err != nil {

		log.Printf("Error making http request %s", err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Printf("Error reading http request %s", err.Error())
		return nil, err
	}

	return body, nil
}

// generateAccessToken sends a http request to generate new access token
func (m *MpesaB2C) generateAccessToken(conn *redis.Client) (*string, error) {

	redisKey := "B2C:ACCESS_TOKEN"

	token, err := GetRedisKey(conn, redisKey)
	if err != nil {

		log.Printf("error getting B2C:ACCESS_TOKEN from redis %s ", err.Error())

	} else {

		if len(token) > 2 {

			return &token, nil
		}
	}

	url := fmt.Sprintf("%s/oauth/v3/generate?grant_type=client_credentials", m.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {

		log.Printf("Error retrieving B2C Token %s", err.Error())
		return nil, err
	}

	req.SetBasicAuth(m.consumerKey, m.consumerSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.makeRequest(req)
	if err != nil {

		log.Printf("Error retrieving B2C Token %s", err.Error())
		return nil, err
	}

	accessTokenResponse := new(MpesaAccessTokenResponseB2C)
	if err := json.Unmarshal(resp, &accessTokenResponse); err != nil {

		log.Printf("Error retrieving B2C Token %s", err.Error())
		return nil, err
	}

	token = accessTokenResponse.AccessToken
	err = SetRedisKeyWithExpiry(conn, redisKey, token, 15*60)
	if err != nil {

		log.Printf("Error saving B2C Token %s", err.Error())

	}

	return &token, nil
}

// setupHttpRequestWithAuth is a helper method aimed to create a http request adding
// the Authorization Bearer header with the access token for the Mpesa app.
func (m *MpesaB2C) setupHttpRequestWithAuth(conn *redis.Client, method, url string, body []byte) (*http.Request, error) {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {

		log.Printf("Error making HTTP request %s", err.Error())
		return nil, err
	}

	accessTokenResponse, err := m.generateAccessToken(conn)
	if err != nil {

		log.Printf("Error generateAccessToken  %s", err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *accessTokenResponse))

	return req, nil
}

// GenerateSecurityCredentials getSecurityCredentials returns the encrypted password using the public key of the specified environment
func GenerateSecurityCredentials(password string, isOnProduction bool) (string, error) {

	path := "/gaming/files/certificates/production.cer"

	if !isOnProduction {

		path = os.Getenv("/gaming/files/certificates/sandbox.cer")

	} else {

		path = os.Getenv("/gaming/files/certificates/production.cer")

	}

	log.Printf("got mpesa certificate path %s ", path)

	f, err := os.Open(path)
	if err != nil {

		log.Printf("error openning mpesa certificate file from %s error %s", path, err.Error())
		return "", err
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	contents, err := ioutil.ReadAll(f)
	if err != nil {

		log.Printf("error reading mpesa certificate file %s", err.Error())
		return "", err
	}

	block, _ := pem.Decode(contents)
	if block == nil {

		log.Printf("error retrieving bytes")
		return "nil", errors.New(fmt.Sprintf("error decoding %s: not a valid PEM encoded block", path))
	}

	var cert *x509.Certificate

	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {

		log.Printf("error processing mpesa certificate file %s", err.Error())
		return "", err
	}

	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	reader := rand.Reader

	encryptedPayload, err := rsa.EncryptPKCS1v15(reader, rsaPublicKey, []byte(password))
	if err != nil {

		log.Printf("error encrypting password %s", err.Error())
		return "", err
	}

	securityCredentials := base64.StdEncoding.EncodeToString(encryptedPayload)
	return securityCredentials, nil

}
