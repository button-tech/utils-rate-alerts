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

func (c *Cache) Set(cur string, f string, b ConditionBlock) {
	c.Lock()
	cValue, ok := c.subscribers[Currency(cur)]
	if !ok {
		c.subscribers[Currency(cur)] = map[Fiat][]ConditionBlock{}
	}

	fValue, _ := cValue[Fiat(f)]
	fValue = append(fValue, b)

	c.Unlock()
}

func (c *Cache) Get() (m map[Currency]map[Fiat][]ConditionBlock) {
	c.Lock()
	m = c.subscribers
	c.Unlock()
	return
}

