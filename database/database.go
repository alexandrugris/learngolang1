package database

import (
	"database/sql"
	"log"
)

// DbConn is our database connection pool
var DbConn *sql.DB

// Connect opens the connection to the database
func Connect() {
	var err error
	DbConn, err = sql.Open("postgres", "user=postgres dbname=products password=mysecretpassword sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}
}
