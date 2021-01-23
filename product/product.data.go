package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"alexandrugris.ro/webservicelearning/database"
)

// Map interface
type Map interface {
	GetAll() []*Product
	FindByID(int) *Product
	DeleteByID(int)
	UpdateByID(int, *Product) bool
	CreateNew(*Product)
}

// internal
type mapInternal struct {
}

var productMap *mapInternal

func (m *mapInternal) GetAll() []*Product {

	// this will allow the queries to timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := database.DbConn.QueryContext(ctx, `
	SELECT 
		ProductID, 
		Manufacturer, 
		PricePerUnit, 
		UnitsAvailable, 
		ProductName 
	FROM Products
	`)

	if err != nil {
		log.Println(err)
		return nil
	}

	defer results.Close()

	ret := make([]*Product, 0, 100)

	for results.Next() {
		v := Product{}

		results.Scan(
			&v.ProductID,
			&v.Manufacturer,
			&v.PricePerUnit,
			&v.UnitsAvailable,
			&v.ProductName)

		ret = append(ret, &v)
	}

	return ret
}

func (m *mapInternal) FindByID(id int) *Product {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	row := database.DbConn.QueryRowContext(ctx, `
	SELECT 
		ProductID, 
		Manufacturer, 
		PricePerUnit, 
		UnitsAvailable, 
		ProductName 
	FROM Products WHERE ProductID=$1`, id)

	v := Product{}

	err := row.Scan(
		&v.ProductID,
		&v.Manufacturer,
		&v.PricePerUnit,
		&v.UnitsAvailable,
		&v.ProductName)

	switch err {
	case sql.ErrNoRows:
		return nil
	case nil:
		return &v
	default:
		log.Panic(err)
	}

	return nil
}

func (m *mapInternal) DeleteByID(id int) {

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := database.DbConn.ExecContext(ctx, "DELETE FROM Products WHERE ProductID=$1", id); err != nil {
		log.Println(err)
	}
}

func (m *mapInternal) UpdateByID(id int, p *Product) bool {

	// allow the database to timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := database.DbConn.ExecContext(ctx, `
	UPDATE Products SET 
		Manufacturer=$1
		PricePerUnit=$2
		UnitsAvailable=$3 
		ProductName=$4
	WHERE ProductID=$5
	`, p.Manufacturer,
		p.PricePerUnit,
		p.UnitsAvailable,
		p.ProductName,
		p.ProductID)

	if err != nil {
		log.Println(err)
		return false
	}

	n, err := res.RowsAffected()

	if err != nil {
		log.Println(err)
		return false
	}

	return n == 1
}

func (m *mapInternal) CreateNew(p *Product) {

	// allow the database to timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stmt, err := database.DbConn.PrepareContext(ctx,
		`INSERT INTO Products(Manufacturer, PricePerUnit, UnitsAvailable, ProductName, ProductID) 
		VALUES ($1, $2, $3, $4, nextval('pk_product')) RETURNING ProductID`)

	if stmt == nil || err != nil {
		log.Fatal(err)
	}

	sqlRow := stmt.QueryRowContext(ctx,
		p.Manufacturer, p.PricePerUnit, p.UnitsAvailable, p.ProductName)

	if err := sqlRow.Scan(&p.ProductID); err != nil {
		log.Fatal(err)
	}
}

var productMapCreateMtx sync.Mutex

func initProducts() {

	log.Println("Seeding products")

	for i := 0; i < 10; i++ {
		p := Product{
			Manufacturer:   "Apple",
			PricePerUnit:   fmt.Sprintf("%vEUR", (rand.Int()%10)*100+500),
			UnitsAvailable: rand.Int() % 15,
			ProductName:    "MacBook Pro",
		}

		productMap.CreateNew(&p)

		log.Printf("Product with ID %v created\n", p.ProductID)

	}
}

// InitStorage initializes the storage
func InitStorage() error {

	if database.DbConn == nil {
		return errors.New("Database not opened")
	}

	t := reflect.TypeOf(Product{})

	query := "CREATE TABLE IF NOT EXISTS Products ("

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		query += f.Name + " "

		switch f.Type.Name() {
		case "string":
			query += "varchar (100)"
		default:
			query += f.Type.Name()
		}

		if i+1 < t.NumField() {
			query += ", "
		}

	}
	query += ");"

	log.Println(query)

	if _, err := database.DbConn.Exec(query); err != nil {
		return err
	}

	if _, err := database.DbConn.Exec("DELETE FROM Products"); err != nil {
		return err
	}

	database.DbConn.Exec("ALTER TABLE Products ADD PRIMARY KEY (ProductID)")

	if _, err := database.DbConn.Exec(
		"CREATE SEQUENCE IF NOT EXISTS pk_product CACHE 100 OWNED BY Products.ProductID"); err != nil {
		return err

	}

	initProducts()
	return nil
}

// GetProductMap returns the singleton map
func GetProductMap() Map {
	if productMap == nil {
		productMapCreateMtx.Lock()
		defer productMapCreateMtx.Unlock()

		if productMap == nil {
			productMap = &mapInternal{}
		}
	}

	return productMap
}
