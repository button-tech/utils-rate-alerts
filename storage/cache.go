package storage

import (
	"sync"
)

type Currency string
type Fiat string
type URL string


type Cache struct {
	sync.Mutex
	subscribers map[Currency]map[Fiat]map[URL]ConditionBlock
}

type ConditionBlock struct {
	Currency  string `json:"currency"`
	Price     string `json:"price"`
	Fiat      string `json:"fiat"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
}

func NewCache() *Cache {
	return &Cache{
		subscribers: make(map[Currency]map[Fiat]map[URL]ConditionBlock),
	}
}

func (c *Cache) Set(b ConditionBlock) {
	c.Lock()
	var ok bool

	if ok = c.setCurrency(b); !ok {
		c.subscribers[Currency(b.Currency)] = make(map[Fiat]map[URL]ConditionBlock)
	}

	if ok = c.setFiat(b); !ok {
		c.subscribers[Currency(b.Currency)][Fiat(b.Fiat)] = make(map[URL]ConditionBlock)
	}

	c.setURL(b)

	c.Unlock()
}

func (c *Cache) setCurrency(b ConditionBlock) bool {
	_, ok := c.subscribers[Currency(b.Currency)]
	return ok
}

func (c *Cache) setFiat(b ConditionBlock) bool {
	_, ok := c.subscribers[Currency(b.Currency)][Fiat(b.Fiat)]
	return ok
}

func (c *Cache) setURL(b ConditionBlock) {
	c.subscribers[Currency(b.Currency)][Fiat(b.Fiat)][URL(b.URL)] = b
}

func (c *Cache) Get() (m map[Currency]map[Fiat]map[URL]ConditionBlock) {
	c.Lock()
	m = c.subscribers
	c.Unlock()
	return
}

