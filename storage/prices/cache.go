package prices

type Cache struct {
	prices []*ParsedPrices
}

type ParsedPrices struct {
	Currency string
	Rates map[string]string
}

func NewCache() *Cache {
	return &Cache{
		prices: make([]*ParsedPrices, 0),
	}
}

func (c *Cache) Set(pp []*ParsedPrices) {
	c.prices = pp
}

func (c *Cache) Get() []*ParsedPrices {
	return c.prices
}
