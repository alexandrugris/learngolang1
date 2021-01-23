package database

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"
)

// DbConn is our database connection pool
var DbConn *sql.DB

const ConnectionString = "user=postgres dbname=products password=mysecretpassword sslmode=disable"

// Connect opens the connection to the database
func Connect() {
	var err error
	DbConn, err = sql.Open("postgres", ConnectionString)

	if err != nil {
		log.Fatal(err)
	}

	// configure the connection pool
	DbConn.SetConnMaxLifetime(60 * time.Second)
	DbConn.SetMaxOpenConns(4)
	DbConn.SetMaxIdleConns(4)
}

// ListenForNotifications should be invoked as a goroutine
func ListenForNotifications(event string, notif func(json []byte)) error {

	listener := pq.NewListener(ConnectionString, 1*time.Second, 10*time.Second, func(ev pq.ListenerEventType, err error) {
		log.Println(ev)
		if err != nil {
			log.Println(err)
		}
	})

	if err := listener.Listen(event); err != nil {
		return err
	}

	for {
		select {
		case n := <-listener.Notify:
			go notif([]byte(n.Extra))

		case <-time.After(90 * time.Second):

			log.Println("No events, pinging the connection")
			if err := listener.Ping(); err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
}
