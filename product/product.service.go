package product

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func productHandler(w http.ResponseWriter, r *http.Request) {

	retCode := func(w http.ResponseWriter, r *http.Request) int {

		mp := GetProductMap()
		pathSegments := strings.Split(r.URL.Path, "/products/")

		if len(pathSegments) != 2 {
			return http.StatusBadRequest
		}

		productID, err := strconv.Atoi(pathSegments[len(pathSegments)-1])

		if err != nil {
			return http.StatusBadRequest
		}

		product := GetProductMap().FindByID(productID)

		if product == nil {
			return http.StatusNotFound
		}

		switch r.Method {
		case http.MethodGet:

			jsonStr, err := json.Marshal(product)
			if err != nil {
				return http.StatusInternalServerError
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonStr))

			return http.StatusOK

		case http.MethodPut:

			body, err := ioutil.ReadAll(
				&io.LimitedReader{
					R: r.Body,
					N: 1024})

			if err != nil || json.Unmarshal(body, &product) != nil {
				return http.StatusBadRequest
			}

			// ensure ID stays the same
			product.ProductID = productID
			mp.UpdateByID(productID, product)

			return http.StatusAccepted
		default:
			return http.StatusMethodNotAllowed
		}

	}(w, r)

	log.Println(r.Method, r.URL.Path)
	if retCode != http.StatusOK {
		w.WriteHeader(retCode)
	}
}

func productsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		jsonStr, err := json.Marshal(GetProductMap().GetAll())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(jsonStr))
		}
	case http.MethodPost:
		body, err := ioutil.ReadAll(
			&io.LimitedReader{
				R: r.Body,
				N: 1024})

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		prod := Product{}
		err = json.Unmarshal(body, &prod)

		if err != nil || prod.ProductID != 0 {

			if err == nil {
				err = errors.New("ProductID should be 0 - if you know the ID, use PUT")
			}

			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		GetProductMap().CreateNew(&prod)
		w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusNotImplemented)
	}

}

// GetHTTPHandlers returns the handlers and the associated routes
func GetHTTPHandlers() map[string]http.HandlerFunc {

	return map[string]http.HandlerFunc{
		"/products":  productsHandler,
		"/products/": productHandler,
	}

}
