package main

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/maybets/kra-service/app/database"
	"bitbucket.org/maybets/kra-service/app/kra_v1/controllers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	// Establish a database connection
	db := database.DbInstance()
	defer db.Close()

	if db == nil {
		fmt.Println("Failed to connect to the database")
		os.Exit(1)
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
