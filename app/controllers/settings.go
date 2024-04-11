package controllers

import (
	"bitbucket.org/maybets/kra-service/app/library"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"strings"
)

type Setting struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RedisSettings struct {
	Settings []Setting `json:"settings"`
}

func (a *Controller) SetSettings(c echo.Context) error {

	u := new(RedisSettings)
	if err := c.Bind(u); err != nil {

		return RespondRaw(c, http.StatusInternalServerError, err.Error())

	}

	for _, k := range u.Settings {

		redisKey := fmt.Sprintf("SETTING:%s", strings.ToUpper(k.Name))
		library.SetRedisKey(a.RedisConnection, redisKey, k.Value)
		log.Printf("set settings %s - val %s", redisKey, k.Value)

	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"data":   "settings saved",
	})
}

func (a *Controller) GetSettings(c echo.Context) error {

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"data": map[string]interface{}{
			"KRA_REGISTRATION_STATUS": library.GetSettings(a.DB, a.RedisConnection, "KRA_REGISTRATION_STATUS", "0"),
			"KRA_ENVIRONMENT":         library.GetSettings(a.DB, a.RedisConnection, "KRA_ENVIRONMENT", "0"),
			"KRA_LOGIN":               library.GetSettings(a.DB, a.RedisConnection, "KRA_LOGIN", ""),
			"KRA_EMAIL":               library.GetSettings(a.DB, a.RedisConnection, "KRA_EMAIL", ""),
			"KRA_PASSWORD":            library.GetSettings(a.DB, a.RedisConnection, "KRA_PASSWORD", ""),
			"B2C_CONSUMER_KEY":        library.GetSettings(a.DB, a.RedisConnection, "B2C_CONSUMER_KEY", ""),
			"B2C_CONSUMER_SECRET":     library.GetSettings(a.DB, a.RedisConnection, "B2C_CONSUMER_SECRET", ""),
			"B2C_INITIATOR":           library.GetSettings(a.DB, a.RedisConnection, "B2C_INITIATOR", ""),
			"B2C_PASSWORD":            library.GetSettings(a.DB, a.RedisConnection, "B2C_PASSWORD", ""),
			"B2C_PAYBILL":             library.GetSettings(a.DB, a.RedisConnection, "B2C_PAYBILL", ""),
		},
	})
}
