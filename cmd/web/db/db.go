package db

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"sync"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
	"xorm.io/xorm/names"
)

var db *xorm.Engine
var createMutex sync.Mutex

func createDb() error {
	createMutex.Lock()
	defer createMutex.Unlock()

	if db != nil {
		return nil
	}

	getConnectionString := func(dbName string) string {
		return fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
			os.Getenv("dbHost"), os.Getenv("dbPort"),
			os.Getenv("dbUser"), os.Getenv("dbPassword"), dbName)
	}

	engine, err := xorm.NewEngine("postgres", getConnectionString("postgres"))

	if err != nil {
		return err
	}

	realDbName := os.Getenv("dbName")

	if engine != nil {
		_, _ = engine.Exec(fmt.Sprintf("CREATE DATABASE %s;", realDbName))

		_ = engine.Close()
	}

	engine, err = xorm.NewEngine("postgres", getConnectionString(realDbName))

	if err != nil {
		return err
	}

	db = engine

	engine.SetSchema("public")
	engine.SetMapper(names.SnakeMapper{})
	engine.SetTZLocation(time.UTC)
	engine.SetTZDatabase(time.UTC)

	engine.ShowSQL(true)

	m := migrate.New(engine, &migrate.Options{
		TableName:    "migrations",
		IDColumnName: "id",
	}, migrations)

	err = m.Migrate()

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func GetDb() (*xorm.Engine, error) {
	if db == nil {
		if err := createDb(); err != nil {
			return nil, err
		}
	}

	return db, nil
}
