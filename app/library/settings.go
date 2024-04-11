package library

import (
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/labstack/gommon/log"
	"strings"
)

func GetSettings(db *sql.DB, redisConn *redis.Client, name string, defaults string) string {

	redisKey := fmt.Sprintf("SETTING:%s", strings.ToUpper(name))

	data, err := GetRedisKey(redisConn, redisKey)
	if err == nil && len(data) > 0 {

		return strings.TrimSpace(data)
	}

	var value sql.NullString

	err = db.QueryRow("SELECT setting_value FROM settings WHERE name = ? ", name).Scan(&value)

	if err != nil {

		log.Printf("got error reading settings %s error %s ", name, err.Error())
		value.String = defaults

	} else {
		if !value.Valid {
			value.String = defaults
		}
	}

	SetRedisKey(redisConn, redisKey, value.String)

	return value.String
}
