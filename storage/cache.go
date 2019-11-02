package storage

import (
	"github.com/pkg/errors"
	"sync"
)

type Token string
type Fiat string
type URL string


type Cache struct {
	sync.Mutex
	subscribers map[Token]map[Fiat]map[URL]ConditionBlock
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
		subscribers: make(map[Token]map[Fiat]map[URL]ConditionBlock),
	}
}

func (c *Cache) Set(b ConditionBlock) {
	c.Lock()
	var ok bool

	if ok = c.setCurrency(b); !ok {
		c.subscribers[Token(b.Currency)] = make(map[Fiat]map[URL]ConditionBlock)
	}

	if ok = c.setFiat(b); !ok {
		c.subscribers[Token(b.Currency)][Fiat(b.Fiat)] = make(map[URL]ConditionBlock)
	}

	c.setURL(b)

	c.Unlock()
}

func (c *Cache) setCurrency(b ConditionBlock) bool {
	_, ok := c.subscribers[Token(b.Currency)]
	return ok
}

func (c *Cache) setFiat(b ConditionBlock) bool {
	_, ok := c.subscribers[Token(b.Currency)][Fiat(b.Fiat)]
	return ok
}

func (c *Cache) setURL(b ConditionBlock) {
	c.subscribers[Token(b.Currency)][Fiat(b.Fiat)][URL(b.URL)] = b
}

func (c *Cache) Get() (m map[Token]map[Fiat]map[URL]ConditionBlock) {
	c.Lock()
	m = c.subscribers
	c.Unlock()
	return
}

func (c *Cache) Delete(b ConditionBlock) error {
	c.Lock()
	_, ok := c.subscribers[Token(b.Currency)][Fiat(b.Fiat)][URL(b.URL)]
	if !ok {
		return errors.New("no key in map")
	}
	delete(c.subscribers[Token(b.Currency)][Fiat(b.Fiat)],  URL(b.URL))
	c.Unlock()
	return nil
}

