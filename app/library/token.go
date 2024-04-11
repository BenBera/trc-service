package library

import (
	"bitbucket.org/maybets/kra-service/app/constants"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/labstack/gommon/log"
	library "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

// decrypt from base64 to decrypted string
func Decrypt(keyString string, stringToDecrypt string) (plainText string, err error) {

	key, _ := hex.DecodeString(keyString)
	ciphertext, _ := base64.URLEncoding.DecodeString(stringToDecrypt)

	block, err := aes.NewCipher(key)
	if err != nil {
		logrus.WithFields(logrus.Fields{constants.DESCRIPTION: "got error creating new cipher block from key"}).Error(err.Error())

		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		logrus.WithFields(logrus.Fields{constants.DESCRIPTION: "ciphertext too short", constants.DATA: ciphertext}).Error(err.Error())

		return "", fmt.Errorf("ciphertext too short")

	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext), nil
}

type LoginRequest struct {
	LoginDetails interface{} `json:"loginDetails"`
}

func DoRegistration(db *sql.DB, redisConn *redis.Client) (error, string) {

	KraRegistrationStatus := os.Getenv("KRA_REGISTRATION_STATUS")
	KraRegistrationStatusInt, _ := strconv.ParseInt(KraRegistrationStatus, 10, 64)
	if KraRegistrationStatusInt == 1 {

		return nil, ""
	}

	KraEnvironment := os.Getenv("KRA_ENVIRONMENT")

	var endpoint string

	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)
	if KraEnvironmentInt == 1 {

		endpoint = os.Getenv("kra_prod_register_url")

	} else {

		endpoint = os.Getenv("kra_test_register_url")

	}

	login := os.Getenv("KRA_LOGIN")
	if KraRegistrationStatus == "0" {
		login = strings.ToLower(login)

	}
	email := os.Getenv("KRA_EMAIL")
	password := os.Getenv("KRA_PASSWORD")
	log.Printf(fmt.Sprintf("Here is login in lowerCase %s", login))

	payload := map[string]interface{}{
		"email":    email,
		"login":    login,
		"password": password,
		"langKey":  "en",
	}
	log.Printf(fmt.Sprintf("Registration Payload %s", payload))

	st, body := library.HTTPPost(endpoint, nil, payload)
	log.Printf(fmt.Sprintf("STATUS HERE =====>>%d", st))

	if st != 201 {
		log.Printf(fmt.Sprintf("Got invalid status during KRA Register\nemail %s\nlogin %s\npassword %s\nstatusCode %d,response%s", email, login, password, st, body))

		//library.Notification(db, cfg, redisConn, "KRA-STATUS", "Failed to Register KRA", fmt.Sprintf("Got invalid status during KRA Register\nemail %s\nlogin %s\npassword %s\nstatusCode %d,response%s", email, login, password, st, body))

	} else {

		redisKey := fmt.Sprintf("SETTING:%s", strings.ToUpper("KRA_REGISTRATION_STATUS"))
		err := SetRedisKey(redisConn, redisKey, "1")
		if err != nil {
			log.Printf(fmt.Sprintf("Failed to save registration status in the redis %s", err.Error()))
		}

		//library.Notification(db, cfg, redisConn, "KRA-STATUS", "Successfully Registered KRA", fmt.Sprintf("email %s\nlogin %s\npassword %s\nstatusCode %d,response%s", email, login, password, st, body))

	}
	return nil, "Company Successfully registered"
}

func BearerToken(db *sql.DB, redisConn *redis.Client) (error, string) {

	KraRegistrationStatus := os.Getenv("KRA_REGISTRATION_STATUS")
	KraRegistrationStatusInt, _ := strconv.ParseInt(KraRegistrationStatus, 10, 64)
	if KraRegistrationStatusInt == 0 {

		return fmt.Errorf("kra is not yet setup"), ""
	}

	KraEnvironment := os.Getenv("KRA_ENVIRONMENT")

	var endpoint string

	KraEnvironmentInt, _ := strconv.ParseInt(KraEnvironment, 10, 64)
	if KraEnvironmentInt == 1 {

		endpoint = os.Getenv("kra_prod_access_token_url")

	} else {

		endpoint = os.Getenv("kra_test_access_token_url")

	}

	login := GetSettings(db, redisConn, "KRA_LOGIN", "")
	password := GetSettings(db, redisConn, "KRA_PASSWORD", "")

	payload := map[string]interface{}{
		"username": strings.ToLower(login),
		"password": password,
	}

	//loginRequest := LoginRequest{LoginDetails: payload}

	var response map[string]interface{}

	st, body := library.HTTPPost(endpoint, nil, payload)

	_ = json.Unmarshal([]byte(body), &response)

	tokenID, err := library.GetString(response, "id_token", "")
	if err != nil {

		log.Printf("error decoding token %s ", err.Error())
	}

	if len(tokenID) == 0 {
		log.Printf(fmt.Sprintf("Got invalid status during KRA Register\nusername %s\npassword %s\nstatusCode %d,response%s", login, password, st, body))

		//library.Notification(db, cfg, redisConn, "KRA-STATUS", "Failed to generate KRA Token", fmt.Sprintf("Got invalid status during KRA Register\nusername %s\npassword %s\nstatusCode %d,response%s", login, password, st, body))
		return fmt.Errorf(body), ""
	}

	return nil, tokenID

}
