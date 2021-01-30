package main

import (
	"alexandrugris.ro/webservicelearning/database"
	"alexandrugris.ro/webservicelearning/product"
	"log"
	"net/http"
	"os"

	// import the postgres database driver
	_ "github.com/lib/pq"
)

func corsMiddleware(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// before the handler
		// add the cors middleware headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			// the pre-flight request, make sure it is handled
			return
		}

		// the actual handler
		handler.ServeHTTP(w, r)

		// after handler
	})
}

func main() {

	log.Println("Service started")

	database.Connect()

	for _, v := range os.Args[1:] {
		switch v {
		case "--dbinit":
			log.Println("Initializing database")
			if err := product.InitStorage(); err != nil {
				log.Fatal(err)
			}
		}
	}

	for k, v := range product.GetHTTPHandlers() {
		http.Handle(k, corsMiddleware(v))
	}

	go func() {
		if err := database.ListenForNotifications("product_change", product.HandleChangeProductNotification); err != nil {
			log.Fatal(err)
		}
	}()

	if err := http.ListenAndServeTLS(":8080", "./certs/cert.pem", "./certs/key.pem", nil); err != nil {
		log.Fatal(err)
	}

}
