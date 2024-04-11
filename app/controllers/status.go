package controllers

import "github.com/labstack/echo/v4"

func (a *Controller) Status(c echo.Context) error {

	return c.JSON(200, "ok")
}
