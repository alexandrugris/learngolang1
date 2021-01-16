package main

import (
	"log"
	"net/http"

	"alexandrugris.ro/webservicelearning/product"
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

	// handler for GET all and POST
	for k, v := range product.GetHTTPHandlers() {
		http.Handle(k, corsMiddleware(v))
	}

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
