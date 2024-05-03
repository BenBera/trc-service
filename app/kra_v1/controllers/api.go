package controllers

import (
	"database/sql"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
	"gopkg.in/ini.v1"
)


type Api struct {
	//Conn rabbitMQ connection
	Conn *amqp.Connection
	//DB database connection
	DB *sql.DB
	//ArriveTime when the request arrived, use to measure api response time
	ArriveTime int64
	//RedisConnection redis connection
	RedisConnection *redis.Client
	//Config configuration file
	Config *ini.File
	//E echo server
	E *echo.Echo
}

func (a *Api) Status(c echo.Context) error {

	return c.JSON(200, "ok")
}

func (a *Api) Version(c echo.Context) error {
	return RespondRaw(c, http.StatusOK, Version{
		Version: "1.0.1",
	})
}