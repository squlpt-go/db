package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type status string

const (
	active   = status("active")
	inactive = status("inactive")
)

type Parent struct {
	*Entity
	ID        int64            `field:"parent_id" primary:"parents"`
	Name      Nullable[string] `field:"parent_name"`
	Status    status           `field:"parent_status"`
	Timestamp time.Time
	Data      Json[map[string]any] `field:"parent_data"`
	DontSave  string               `field:"no_field"`
}

func (p *Parent) Hydrate(fields map[string]any) error {
	if name, ok := GetField[string](fields, "hydrate_name"); ok {
		return p.Name.Scan(name)
	}

	if timestamp, ok := GetField[time.Time](fields, "parent_timestamp"); ok {
		p.Timestamp = timestamp
	}

	return nil
}

type Child struct {
	*Entity
	Parent Parent `field:"parent_id" foreign:"parents"`
	ID     int64  `field:"child_id" primary:"children"`
	Name   string `field:"child_name"`
}

type Friend struct {
	*Entity
	ID   int64  `field:"friend_id" primary:"friends"`
	Name string `field:"friend_name"`
}

func MustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)

	if !ok {
		panic("env var " + key + " must be set")
	}

	return v
}

var _db *sql.DB

func DB() *sql.DB {
	if _db == nil {
		err := godotenv.Load(".env.test")

		if err != nil {
			panic(err)
		}

		_db, err = sql.Open("mysql", MustGetEnv("DB_DSN")+"?parseTime=true&multiStatements=true")

		if err != nil {
			panic(err)
		}

		err = _db.Ping()

		if err != nil {
			panic(err)
		}

		path := filepath.Join("sql", "test_mysql.sql")
		fmt.Printf("Importing %s\n", path)

		qs, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		_, err = _db.Exec(string(qs))
		if err != nil {
			panic(err)
		}

		time.Sleep(3 * time.Second)
	}

	return _db
}
