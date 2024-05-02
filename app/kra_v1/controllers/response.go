package controllers

import (
	"encoding/json"

	"bitbucket.org/maybets/kra-service/app/models"
	"github.com/labstack/echo/v4"
)

// RespondJSON makes the response with payload as json format
func response(e echo.Context, status int, payload []byte) error {

	e.Response().Header().Set("Access-Control-Allow-Origin","*");
	e.Response().Header().Set("Content-Type", "application/json; charset=UTF-8")
	e.Response().WriteHeader(status)
	return e.String(status,string(payload))
}

// respondError makes the error response with payload as json format
func Response(e echo.Context, code int, message models.ResponseMessage) error {

	msg, _ := json.Marshal(message)
	return response(e, code, msg)
}

// respondError makes the error response with payload as json format
func RespondJSON(e echo.Context, code int, message interface{}) error {

	res := models.ResponseMessage{}
	res.Status = code
	res.Message = message
	msg, _ := json.Marshal(res)
	return response(e, code, msg)
}

// respondError makes the error response with payload as json format
func RespondRaw(e echo.Context, code int, message interface{}) error {

	msg, _ := json.Marshal(message)
	return response(e, code, msg)
}