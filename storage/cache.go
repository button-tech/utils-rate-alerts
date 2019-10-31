package storage

import (
	"sync"
)

type Currency string
type Fiat string

type Cache struct {
	sync.Mutex
	subscribers map[Currency]map[Fiat][]ConditionBlock
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
		subscribers: make(map[Currency]map[Fiat][]ConditionBlock),
	}
}

func (c *Cache) Set(b ConditionBlock) {
	c.Lock()
	_, ok := c.subscribers[Currency(b.Currency)]
	if !ok {
		c.subscribers[Currency(b.Currency)] = map[Fiat][]ConditionBlock{Fiat(b.Fiat):make([]ConditionBlock, 0)}
	}

	fValue, ok := c.subscribers[Currency(b.Currency)][Fiat(b.Fiat)]
	fValue = append(fValue, b)
	c.subscribers[Currency(b.Currency)][Fiat(b.Fiat)] = fValue
	c.Unlock()
}

func (c *Cache) Get() (m map[Currency]map[Fiat][]ConditionBlock) {
	c.Lock()
	m = c.subscribers
	c.Unlock()
	return
}

