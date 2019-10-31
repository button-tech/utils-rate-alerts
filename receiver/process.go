package receiver

import (
	"encoding/json"
	"strconv"
	"strings"

	"log"
	"os"
	"time"

	"github.com/button-tech/rate-alerts/storage"
	"github.com/button-tech/rate-alerts/storage/prices"

	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/valyala/fastjson"
)

type requestBlocks struct {
	Tokens []string `json:"tokens"`
	Currencies     []string `json:"currencies"`
	API      string   `json:"api"`
}

func (r *Receiver) deliveryChannel() (<-chan amqp.Delivery, error) {
	msgs, err := r.rabbitMQ.Channel.Consume(
		r.rabbitMQ.Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (r *Receiver) Processing() {
	c, err := r.deliveryChannel()
	if err != nil {
		log.Println(err)
		return
	}

	for msg := range c {
		var block storage.ConditionBlock
		if err := json.Unmarshal(msg.Body, &block); err != nil {
			log.Println(err)
		}
		r.store.Set(block)
	}

	select {

	}
}

func (r *Receiver) GetPrices() {
	t := time.NewTicker(time.Second * 15)
	for ; ; <-t.C {
		var blocks requestBlocks
		blocks.API = "cmc"

		checkMap := make(map[string]struct{})
		m := r.store.Get()
		for currency, fiat := range m {
			blocks.Tokens = append(blocks.Tokens, string(currency))
			for f := range fiat {
				if _, ok := checkMap[string(f)]; !ok {
					checkMap = map[string]struct{}{}
				}
				checkMap[string(f)] = struct{}{}
			}
		}
		
		for k := range checkMap {
			blocks.Currencies = append(blocks.Currencies, k)
		}

		if err := r.getAndStorePrices(&blocks); err != nil {
			log.Println(err)
		}
	}
}

func (r *Receiver) getAndStorePrices(b *requestBlocks) error {
	gotPrices, err := doRequest(b)
	if err != nil {
		return err
	}
	r.schedule(gotPrices)

	return nil
}

func doRequest(b *requestBlocks) ([]*prices.ParsedPrices, error) {
	rq := req.New()
	resp, err := rq.Post(os.Getenv("PRICES"), req.BodyJSON(&b))
	if err != nil {
		return nil, err
	}
	
	ps, err := respFastJSON(resp.Bytes())
	if err != nil {
	return nil, err
	}

	return ps, nil
}

// todo: OPTIMIZATION
func (r *Receiver) schedule(pp []*prices.ParsedPrices) {
	m := r.store.Get()
	var requests []string

	for currency, fiat := range m {
		for f, condition := range fiat {
			for _, p := range pp {
				if string(f) == p.Currency {
					for token, price := range p.Rates {
						if token == string(currency) {
							for _, c := range condition {
								currentPrice, err := strconv.ParseFloat(price, 64)
								conditionPrice, err := strconv.ParseFloat(c.Price, 64)
								if err != nil {
									log.Println(err)
									return
								}
								if c.Condition == "==" && currentPrice == conditionPrice {
									requests = append(requests, c.URL)
								} else if c.Condition == ">=" && currentPrice >= conditionPrice {
									requests = append(requests, c.URL)
								} else if c.Condition == "<=" && currentPrice <= conditionPrice {
									requests = append(requests, c.URL)
								}
							}
						}
					}
				}
			}
		}
	}

	if len(requests) > 0 {
		for _, url := range requests {
			go hunting202(url)
		}
	}
}

type parsedPrices struct {
	currency string
	rates map[string]string
}

func respFastJSON(b []byte) ([]*prices.ParsedPrices, error) {
	var p fastjson.Parser
	parsed, err := p.ParseBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "parseBytes")
	}

	var pp []*prices.ParsedPrices

	o := parsed.GetObject()
	data := o.Get("data")
	array, err := data.Array()
	if err != nil {
		return nil, errors.Wrap(err, "can't get array")
	}
	for _, v := range array {
		obj, err := v.Object()
		if err != nil {
			return nil, errors.Wrap(err, "can't get object")
		}

		var p prices.ParsedPrices
		m := make(map[string]string)
		obj.Visit(func(key []byte, v *fastjson.Value) {
			sKey := string(key)
			if sKey == "currency" {
				c := v.String()
				c = strings.TrimPrefix(c, "\"")
				c = strings.TrimSuffix(c, "\"")
				p.Currency = c
			}

			if sKey == "rates" {
				rates, _ := v.Array()
				for _, r := range rates {
					rO, _ := r.Object()
					rO.Visit(func(key []byte, v *fastjson.Value) {
						r := v.String()
						r = strings.TrimPrefix(r, "\"")
						r = strings.TrimSuffix(r, "\"")
						m[string(key)] = r
					})
				}
				p.Rates = m
			}
		})
		pp = append(pp, &p)
	}

	return pp, nil
}

func hunting202(url string) {
	var err error
	t := time.NewTicker(time.Second * 3)

	counter := 0
	for ; counter != 3; <-t.C {
		err = checkURL(url)
		if err == nil {
			break
		}
		log.Println(err)
		counter++
	}

	if err != nil {
		log.Println(err)
	}
}

func checkURL(url string) error {
	rq := req.New()
	resp, err := rq.Get(url)
	if err != nil {
		return errors.Wrap(err, "checkURL")
	}

	if resp.Response().StatusCode != 202 {
		return errors.Wrap(errors.New("response statusCode not 202"), "checkURL")
	}

	return nil
}
