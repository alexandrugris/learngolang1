package product

import (
	"fmt"
	"math/rand"
	"sync"
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
	mtx sync.RWMutex
	mp  map[int]*Product
}

var productMap *mapInternal

func (m *mapInternal) GetAll() []*Product {

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	ret := make([]*Product, 0, len(m.mp))

	for _, v := range m.mp {
		ret = append(ret, v.CloneTo(&Product{}))
	}

	return ret
}

func (m *mapInternal) FindByID(id int) *Product {

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	if v, found := m.mp[id]; found {
		return v.CloneTo(&Product{})
	}

	return nil
}

func (m *mapInternal) DeleteByID(id int) {

	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.mp, id)
}

func (m *mapInternal) UpdateByID(id int, p *Product) bool {

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if v, found := m.mp[id]; found {
		p.CloneTo(v)
		return true
	}
	return false
}

func (m *mapInternal) CreateNew(p *Product) {

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.mp[p.ProductID] = p.CloneTo(&Product{})
}

var productMapCreateMtx sync.Mutex

func initProducts() {

	for i := 0; i < 10; i++ {
		productMap.mp[i] = &Product{
			ProductID:      i,
			Manufacturer:   "Apple",
			PricePerUnit:   fmt.Sprintf("%vEUR", (rand.Int()%10)*100+500),
			UnitsAvailable: rand.Int() % 15,
			ProductName:    "MacBook Pro",
		}
	}
}

// GetProductMap returns the singleton map
func GetProductMap() Map {

	if productMap == nil {
		productMapCreateMtx.Lock()
		defer productMapCreateMtx.Unlock()

		if productMap == nil {
			productMap = &mapInternal{
				mtx: sync.RWMutex{},
				mp:  make(map[int]*Product),
			}

			initProducts()

		}
	}

	return productMap
}
