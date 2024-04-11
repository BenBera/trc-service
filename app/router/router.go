package router

import (
	"bitbucket.org/maybets/kra-service/app/controllers"
	"bitbucket.org/maybets/kra-service/app/crontask"
	db "bitbucket.org/maybets/kra-service/app/database"

	"bitbucket.org/maybets/kra-service/app/queue"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"log"
	"net/http"
	"os"
	"time"
)

// router and DB instance
type App struct {
	E          *echo.Echo
	DB         *sql.DB
	RedisConn  *redis.Client
	Controller *controllers.Controller
	//betting.UnimplementedBettingServer
}

//var defaultConfigPath = "/gaming/application/config/config.ini"

// Initialize initializes the app with predefined configuration
func (a *App) Initialize() {

	// get rabbitMQ connetion
	a.DB = db.DbInstance()
	a.RedisConn = db.RedisClient()

	cronjob := crontask.Crontask{
		RabbitMQConn: db.GetRabbitMQConnection(),
		DB:           a.DB,
		RedisConn:    a.RedisConn,
		//WalletServiceClient: NewWalletServiceClient(os.Getenv("wallet_service_endpoint")),
		Tracer: otel.Tracer("crontask"),
	}

	go cronjob.SetupJobs()

	go a.setRouters()

	q := queue.Queue{
		ConsumerConnection:   db.GetRabbitMQConnection(),
		PublisherQConnection: db.GetRabbitMQConnection(),
		DB:                   a.DB,
		RedisConn:            db.RedisClient(),
		//WalletServiceClient:   NewWalletServiceClient(os.Getenv("wallet_service_endpoint")),
		//IdentityServiceClient: NewIdentityServiceClient(os.Getenv("identity_service_endpoint")),
		Tracer: otel.Tracer("queue"),
	}

	go q.InitQueues()

}

// setRouters sets the all required router
func (a *App) setRouters() {

	controller := controllers.Controller{
		Conn:            db.GetRabbitMQConnection(),
		DB:              a.DB,
		RedisConnection: a.RedisConn,
		//OddsServiceClient:     NewOddServiceClient(os.Getenv("odds_service_endpoint")),
		//IdentityServiceClient: NewIdentityServiceClient(os.Getenv("identity_service_endpoint")),
		//FixtureServiceClient:  NewFixtureServiceClient(os.Getenv("fixture_service_endpoint")),
		//WalletServiceClient:   NewWalletServiceClient(os.Getenv("wallet_service_endpoint")),
		//BonusServiceClient:    NewBonusServiceClient(os.Getenv("bonus_service_endpoint")),
		Tracer: otel.Tracer("rest-api"),
	}

	a.Controller = &controller

	// init webserver
	a.E = echo.New()
	a.E.Static("/doc", "api")

	a.E.Use(middleware.Gzip())
	a.E.IPExtractor = echo.ExtractIPFromXFFHeader()
	// add recovery middleware to make the system null safe
	a.E.Use(middleware.Recover()) // change due to swagger
	a.E.Use(session.Middleware(sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))))

	a.E.Use(middleware.RequestID())

	uptrace.ConfigureOpentelemetry(
		// copy your project DSN here or use UPTRACE_DSN env var
		uptrace.WithDSN(os.Getenv("UPTRACE_DSN")),

		uptrace.WithServiceName("betting-service"),
		uptrace.WithServiceVersion("v1.0.0"),
		uptrace.WithDeploymentEnvironment("production"),
	)

	a.E.Use(otelecho.Middleware("betting-service"))

	// setup log format and parameters to log for every request

	// Instrument logrus.
	logrus.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	)))

	a.E.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {

			req := c.Request()
			res := c.Response()
			start := values.StartTime
			startMicro := start.UnixMicro()

			stop := time.Now()
			stopMicro := stop.UnixMicro()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {

				id = res.Header().Get(echo.HeaderXRequestID)
			}

			reqSize := req.Header.Get(echo.HeaderContentLength)
			if reqSize == "" {

				reqSize = "0"
			}

			traceID := req.Header.Get("trace-id")
			if traceID == "" {

				traceID = "0"
			}

			service, _ := os.Hostname()

			logrus.WithContext(c.Request().Context()).WithFields(logrus.Fields{
				"service":  service,
				"id":       id,
				"ip":       c.RealIP(),
				"time":     stop.Format(time.RFC3339),
				"host":     req.Host,
				"method":   req.Method,
				"uri":      req.RequestURI,
				"status":   res.Status,
				"size":     reqSize,
				"referer":  req.Referer(),
				"ua":       req.UserAgent(),
				"ttl":      stopMicro - startMicro,
				"trace-id": traceID,
			}).Info("API Response")

			return nil
		},
	}))

	allowedMethods := []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions}
	AllowOrigins := []string{"*"}

	//setup CORS
	corsConfig := middleware.CORSConfig{
		AllowOrigins: AllowOrigins, // in production limit this to only known hosts
		AllowHeaders: AllowOrigins,
		AllowMethods: allowedMethods,
	}

	a.E.Use(middleware.CORSWithConfig(corsConfig))

	// public
	a.E.POST("/settings", a.CreateSettings)
	a.E.PATCH("/update", a.CreateSettings)
	a.E.GET("/fetch", a.GetSettings)

	//status
	a.E.POST("/", a.GetStatus)
	a.E.GET("/", a.GetStatus)

	//swagger
	a.E.GET("/docs/*", echoSwagger.WrapHandler)

}

// Run the app on it's router
func (a *App) Run() {

	server := fmt.Sprintf("%s:%s", os.Getenv("system_host"), os.Getenv("system_port"))
	log.Printf(" HTTP listening on %s ", server)
	a.E.Logger.Fatal(a.E.Start(server))
}
