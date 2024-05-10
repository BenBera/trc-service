package models

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {

	//ErrorCode error status code
	ErrorCode int `json:"error_code"  validate:"required"`

	//ErrorMessage status description
	ErrorMessage string `json:"error_message"  validate:"required"`
}

type SuccessResponse struct {
	//Status http status code
	Status int `json:"status"  validate:"required"`

	//Message status description
	Message string `json:"data"  validate:"required"`
}

type SuccessResponseWithID struct {
	//Status http status code
	Status int `json:"status"  validate:"required"`

	//ID created resource ID
	ID int64 `json:"id"  validate:"required"`

	//Message status description
	Message string `json:"data"  validate:"required"`
}

type BetSuccessResponse struct {
	//Status http status code
	Status int `json:"status"  validate:"required"`

	//BetID bet id that was created
	BetID string `json:"bet_id"  validate:"required"`

	//Message status description
	Message string `json:"data"  validate:"required"`
}

type ResponseMessage struct {
	Status  int         `json:"status"  validate:"required"`
	Message interface{} `json:"message"  validate:"required"`
}

// respondError makes the error response with payload as json format
func RespondRaw(e echo.Context, code int, message interface{}) error {

	msg, _ := json.Marshal(message)
	return response(e, code, msg)
}


// RespondJSON makes the response with payload as json format
func response(e echo.Context, status int, payload []byte) error {

	e.Response().Header().Set("Access-Control-Allow-Origin","*");
	e.Response().Header().Set("Content-Type", "application/json; charset=UTF-8")
	e.Response().WriteHeader(status)
	return e.String(status,string(payload))
}
