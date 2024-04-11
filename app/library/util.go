package library

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

func MysqlToday() string {

	return time.Now().Format("2006-01-02")

}
func GetTime() int64 {

	return int64(time.Now().UnixNano() / 1000000)
}

func ToBase64Token(int642 string) string {

	return base64.StdEncoding.EncodeToString([]byte(int642))

}
func TransactionDate() string {

	return time.Now().Format("2006-01-02T15:04:05")
}

func GetCheckSum(val string) string {

	return fmt.Sprintf("%x", sha256.Sum256([]byte(val)))

}
