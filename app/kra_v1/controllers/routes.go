package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(e *echo.Echo, a *Api, db *sql.DB) {

	// init webserver
	a.E = echo.New()
	
	a.E.IPExtractor = echo.ExtractIPFromRealIPHeader()

	a.E.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${status} latency=${latency_human} - ${uri} - ip=${remote_ip} \n",
		Output: log.Default().Writer(),
	}))

	//setup CORS
	a.E.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // in production limit this to only known hosts
		AllowHeaders: []string{"*"},
		//AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderXForwardedFor,echo.HeaderXRealIP,echo.HeaderAuthorization},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// Rate Limiter
	rconfig := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(

			middleware.RateLimiterMemoryStoreConfig{Rate: 100, Burst: 60, ExpiresIn: 1 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}

	a.E.Use(middleware.RateLimiterWithConfig(rconfig))

	a.E.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
		OnTimeoutRouteErrorHandler: func(err error, e echo.Context) {
		},
	}))

	//all routes
	e.POST("/status", a.Status)
	e.POST("/version", a.Version)

	e.POST("/kqra-dash-data", func(c echo.Context) error {
		// Call a.GetKqraDashData with the necessary parameters
		err := a.GetKqraDashData(c, db, a.RedisConnection)
		if err != nil {
			// Handle error if needed
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Failed to process request",
			})
		}
	
		// Return success response if no error occurred
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Request processed successfully",
		})
	})

}
