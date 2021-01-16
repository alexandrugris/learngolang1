package product

// Product type
type Product struct {
	ProductID      int    `json:"productId"`
	Manufacturer   string `json:"manufacturer"`
	PricePerUnit   string `json:"pricePerUnit"`
	UnitsAvailable int    `json:"unitsAvailable"`
	ProductName    string `json:"productName"`
}

// CloneTo clones this product to another
func (p *Product) CloneTo(p2 *Product) *Product {
	p2.ProductID = p.ProductID
	p2.Manufacturer = p.Manufacturer
	p2.PricePerUnit = p.PricePerUnit
	p2.UnitsAvailable = p.UnitsAvailable
	p2.ProductName = p.ProductName
	return p2
}
