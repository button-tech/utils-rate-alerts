package bot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type cache struct {
	mu          sync.Mutex
	subscribers map[string][]page
	language    map[string]string
	alerts      map[string][]string
}

func newCache() *cache {
	return &cache{
		mu:          sync.Mutex{},
		subscribers: make(map[string][]page),
		language:    make(map[string]string),
		alerts:      make(map[string][]string),
	}
}

func (c *cache) setAlert(chatID int64, currency, fiat, price, condition string) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	val, ok := c.alerts[k]
	if !ok {
		c.alerts[k] = make([]string, 0)
	}
	val = append(val, genAlertValue(currency, fiat, price, condition))
	c.alerts[k] = val
	c.mu.Unlock()
}

func genAlertValue(currency, fiat, price, condition string) string {
	return fmt.Sprintf("%s_%s_%s_%s", currency, fiat, price, condition)
}

func (c *cache) setRawAlerts(chatID int64, alerts []string) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	c.alerts[k] = alerts
	c.mu.Unlock()
}

func (c *cache) getRawAlerts(chatID int64) ([]string, bool) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.alerts[k]
	if !ok {
		return nil, ok
	}
	return val, true
}

func (c *cache) getAlerts(chatID int64) (a string, ok bool) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.alerts[k]
	if !ok {
		return "", ok
	}

	for i, v := range val {
		alertItems := strings.Split(v, "_")
		u := userAlert{
			currency:  alertItems[0],
			fiat:      alertItems[1],
			price:     alertItems[2],
			condition: alertItems[3],
		}
		one := fmt.Sprintf(
			"%s %s %s %s",
			strings.ToUpper(u.currency),
			u.condition,
			u.price,
			strings.ToUpper(u.fiat),
		)
		a += fmt.Sprintf("№%s %s \n", strconv.Itoa(i+1), one)
	}
	return
}

func (c *cache) deleteAlert(chatID int64, alert string) (string, bool) {
	var deleted string
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.alerts[k]
	if !ok {
		return "", true
	}

	numb, err := strconv.Atoi(alert)
	if err != nil {
		return "", false
	}

	if len(val) >= numb {
		deleted = val[numb-1]
		val = append(val[:numb-1], val[numb-1+1:]...)
		c.alerts[k] = val
		return deleted, true
	}

	return "", false
}

func (c *cache) checkLanguage(username string) (bool, string) {
	c.mu.Lock()
	l, ok := c.language[username]
	c.mu.Unlock()
	return ok, l
}

func (c *cache) setupLanguage(username, language string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.language[username] = language
	return language
}

func (c *cache) set(k int64, p page) []page {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := keyGen(k)
	val, ok := c.subscribers[key]
	if !ok {
		c.subscribers[key] = make([]page, 0, 4)
	}
	val = append(val, p)
	c.subscribers[key] = val
	return val
}

func (c *cache) get(k int64) (ps []page, ok bool) {
	key := keyGen(k)
	if ps, ok = c.subscribers[key]; !ok {
		return nil, false
	}
	return
}

func (c *cache) back(k int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := keyGen(k)
	ps := c.subscribers[key]
	if len(ps) <= 1 {
		delete(c.subscribers, key)
		return
	}
	c.subscribers[key] = ps[:len(ps)-1]
}

func (c *cache) delete(k int64) {
	c.mu.Lock()
	key := keyGen(k)
	delete(c.subscribers, key)
	c.mu.Unlock()
}
