package router

import "github.com/labstack/echo/v4"

//	CreateSettings create bet settings
//
// @Summary Create Wallet Settings
// @Description This API creates bet settings,
// @Tags admin - deposit settings
// @Security ApiKeyAuth
// @Param request body models.Settings true "Settings Details"
// @Accept json
// @Produce json
// @Success      201  {object}  models.SuccessResponse "Settings created successfully"
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /setting [post]
func (a *App) CreateSettings(c echo.Context) error {

	return a.Controller.SetSettings(c)
}

//	GetSettings get wallet settings
//
// @Summary Get wallet Settings
// @Description This API gets bet settings,
// @Tags admin - deposit settings
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success      200  {object}  models.Settings "Settings "
// @Failure      400  {object}  models.ErrorResponse "Invalid payload or missing field in the payload"
// @Failure      401  {object}  models.ErrorResponse "Authorization error"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router /setting [get]
func (a *App) GetSettings(c echo.Context) error {

	return a.Controller.GetSettings(c)
}
