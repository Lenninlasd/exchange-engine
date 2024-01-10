package postgres

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

const (
	user   = "postgres"
	dbname = "engine"
)

var host = os.Getenv("POSTGRES_HOST")
var password = os.Getenv("POSTGRES_PASS")
var port = os.Getenv("POSTGRES_PORT")

func OpenConnection() *sql.DB {
	psqlconn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// check db
	err = db.Ping()
	CheckError(err)

	return db
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
