package storage

import (
	"github.com/pkg/errors"
	"sync"

	"github.com/google/uuid"
)

type Cache struct {
	sync.Mutex
	mapIdentifier map[string]string
	mapData       map[string]data
}

type data struct {
	currency  string
	price     string
	fiat      string
	condition string
}

func NewCache() *Cache {
	return &Cache{
		mapIdentifier: make(map[string]string),
		mapData:       make(map[string]data),
	}
}

func (c *Cache) Set(url string) error {
	c.Lock()
	u, err := uuidGen()
	if err != nil {
		return err
	}
	c.mapIdentifier[u] = url

	c.Unlock()
	return nil
}

func (c *Cache) Get() (m map[string]string) {
	c.Lock()
	m = c.mapIdentifier
	c.Unlock()
	return
}

func uuidGen() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "UUID generate")
	}
	return u.String(), nil
}
