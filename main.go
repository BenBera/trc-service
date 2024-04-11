package main

import (
	"bitbucket.org/maybets/kra-service/app/database"
	"bitbucket.org/maybets/kra-service/app/router"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"runtime"
)

func main() {

	//setup database
	dbInstance := database.DbInstance()

	driver, err := mysql.WithInstance(dbInstance, &mysql.Config{})
	if err != nil {

		logrus.Panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file:///%s/migrations", GetRootPath()), "mysql", driver)
	if err != nil {

		logrus.Errorf("migration setup error %s ", err.Error())
	}

	err = m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err != nil {

		logrus.Errorf("migration error %s ", err.Error())
	}

	// setup consumers
	var a router.App
	a.Initialize()
	//go a.GRPC()
	a.Run()

}

func GetRootPath() string {

	_, b, _, _ := runtime.Caller(0)

	// Root folder of this project
	return filepath.Join(filepath.Dir(b), "./")
}
