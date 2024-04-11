package router

import "github.com/labstack/echo/v4"

//	GetStatus get Kra service health
//
// @Summary get Kra service health
// @Description This API gets Kra service health
// @Tags health check
// @Accept json
// @Produce json
// @Success      200  {object}  map[string]interface{} "Status "
// @Failure      500  {object}  models.ErrorResponse "Service not healthy"
// @Router / [get]

func (a *App) GetStatus(c echo.Context) error {

	return a.Controller.Status(c)
}
