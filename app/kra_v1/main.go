package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"bitbucket.org/maybets/kra-service/app/database"
	"bitbucket.org/maybets/kra-service/app/kra_v1/controllers"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {

	// Establish a database connection
	db := database.DbInstance()
	defer db.Close()

	if db == nil {
		fmt.Println("Failed to connect to the database")
		os.Exit(1)
	}


	// Database migration setup
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Failed to initialize database driver: %v",err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file:///%s/migrations", GetRootPath()), "mysql", driver)
	if err != nil {
		logrus.Errorf("migration setup error: %s", err.Error())
	}

	err = m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err != nil {
		logrus.Errorf("migration error: %s", err.Error())
	}

	
	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize your API instance
	api := &controllers.Api{
		E: e,		
	}

	// api.SetupRoutes(e)
	controllers.SetupRoutes(api.E, api, db)

	// Start the server
	port := ":8085"
	log.Printf("Server running on port %s", port)
	log.Fatal(e.Start(port))
}

func GetRootPath() string {

	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	return filepath.Join(filepath.Dir(b), "./")
}

